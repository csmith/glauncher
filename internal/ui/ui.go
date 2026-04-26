package ui

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	xdraw "golang.org/x/image/draw"

	"chameth.com/glauncher/internal/config"
	"chameth.com/glauncher/internal/search"
	"chameth.com/glauncher/internal/x11"
)

const resultHeightDp = 48
const inputHeightDp = 52
const dividerHeightDp = 1
const visibleResults = 8
const windowHeightDp = inputHeightDp + dividerHeightDp + resultHeightDp*visibleResults

type themeConfig struct {
	background color.NRGBA
	divider    color.NRGBA
	primary    color.NRGBA
	secondary  color.NRGBA
	selection  color.NRGBA
	typeface   font.Typeface
}

type App struct {
	providers    []search.Provider
	window       *app.Window
	editor       widget.Editor
	list         layout.List
	results      []search.Result
	selected     int
	query        string
	theme        *material.Theme
	colors       themeConfig
	needsRefresh atomic.Bool
}

func New(themeCfg config.ThemeConfig, providers ...search.Provider) *App {
	colors := parseThemeColors(themeCfg)
	a := &App{
		providers: providers,
		theme:     newTheme(colors),
		colors:    colors,
	}
	a.editor.SingleLine = true
	return a
}

func (a *App) Run() {
	go func() {
		a.window = &app.Window{}
		a.window.Option(
			app.Size(unit.Dp(600), unit.Dp(windowHeightDp)),
			app.Decorated(false),
			app.Title("glauncher"),
		)

		for _, p := range a.providers {
			if ai, ok := p.(search.AsyncInitializer); ok {
				go func() {
					<-ai.Ready()
					a.needsRefresh.Store(true)
					a.window.Invalidate()
				}()
			}
			if asp, ok := p.(search.AsyncSearchProvider); ok {
				asp.SetInvalidate(func() {
					a.needsRefresh.Store(true)
					a.window.Invalidate()
				})
			}
		}

		if err := a.loop(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func (a *App) loop() error {
	var ops op.Ops
	focused := false

	for {
		switch e := a.window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.ConfigEvent:
			if focused && !e.Config.Focused {
				a.window.Perform(system.ActionClose)
			}
			if e.Config.Focused {
				focused = true
			}
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			a.frame(gtx)
			e.Frame(gtx.Ops)
		case app.X11ViewEvent:
			if e.Valid() {
				x11.SetNoDecorations(e.Display, e.Window)
			}
		}
	}
}

func (a *App) frame(gtx layout.Context) {
	gtx.Execute(key.FocusCmd{Tag: &a.editor})

	a.handleNavKeys(gtx)
	a.handleEditorEvents(gtx)
	a.updateSearch()
	a.layout(gtx)
}

func (a *App) handleNavKeys(gtx layout.Context) {
	for {
		ev, ok := gtx.Event(
			key.Filter{Focus: &a.editor, Name: key.NameUpArrow},
			key.Filter{Focus: &a.editor, Name: key.NameDownArrow},
			key.Filter{Focus: &a.editor, Name: key.NamePageUp},
			key.Filter{Focus: &a.editor, Name: key.NamePageDown},
			key.Filter{Focus: &a.editor, Name: key.NameReturn},
			key.Filter{Focus: &a.editor, Name: key.NameEnter},
			key.Filter{Focus: &a.editor, Name: key.NameEscape},
			key.Filter{Focus: &a.editor, Name: key.NameDeleteBackward, Required: key.ModCtrl},
			key.Filter{Focus: &a.editor, Name: key.NameLeftArrow, Required: key.ModCtrl},
			key.Filter{Focus: &a.editor, Name: key.NameRightArrow, Required: key.ModCtrl},
		)
		if !ok {
			break
		}
		ke, ok := ev.(key.Event)
		if !ok || ke.State != key.Press {
			continue
		}

		switch ke.Name {
		case key.NameEscape:
			a.window.Perform(system.ActionClose)
		case key.NameUpArrow:
			if a.selected > 0 {
				a.selected--
				a.scrollToSelected()
			}
		case key.NameDownArrow:
			if a.selected < len(a.results)-1 {
				a.selected++
				a.scrollToSelected()
			}
		case key.NamePageUp:
			a.pageUp()
		case key.NamePageDown:
			a.pageDown()
		case key.NameReturn, key.NameEnter:
			if a.selected < len(a.results) {
				r := a.results[a.selected]
				go func() {
					if err := r.Exec(); err != nil {
						log.Printf("launch error: %v", err)
					}
				}()
				a.window.Perform(system.ActionClose)
			}
		case key.NameDeleteBackward:
			a.deleteWordBack()
		case key.NameLeftArrow:
			a.moveWordLeft()
		case key.NameRightArrow:
			a.moveWordRight()
		}
	}
}

func (a *App) scrollToSelected() {
	if len(a.results) == 0 {
		return
	}
	pos := &a.list.Position
	if pos.Count == 0 {
		return
	}
	if a.selected < pos.First {
		pos.First = a.selected
		pos.Offset = 0
	} else if pos.Count > 1 && a.selected >= pos.First+pos.Count {
		pos.First = a.selected - pos.Count + 1
		pos.Offset = 0
	}
}

func (a *App) pageUp() {
	pos := &a.list.Position
	if pos.Count <= 1 || len(a.results) == 0 {
		return
	}
	firstVisible := pos.First
	if a.selected > firstVisible {
		a.selected = firstVisible
		a.scrollToSelected()
		return
	}
	target := max(firstVisible-pos.Count+1, 0)
	a.selected = target
	pos.First = a.selected
	pos.Offset = 0
}

func (a *App) pageDown() {
	pos := &a.list.Position
	if pos.Count <= 1 || len(a.results) == 0 {
		return
	}
	last := len(a.results) - 1
	lastFullyVisible := min(pos.First+pos.Count-1, last)
	if a.selected < lastFullyVisible {
		a.selected = lastFullyVisible
		a.scrollToSelected()
		return
	}
	target := min(lastFullyVisible+pos.Count-1, last)
	if target <= a.selected {
		target = last
	}
	a.selected = target
	pos.First = max(a.selected-pos.Count+1, 0)
	pos.Offset = 0
}

func (a *App) handleEditorEvents(gtx layout.Context) {
	for {
		_, ok := a.editor.Update(gtx)
		if !ok {
			break
		}
	}
}

func (a *App) updateSearch() {
	q := a.editor.Text()
	if q == a.query && (!a.needsRefresh.Load() || q == "") {
		return
	}
	a.needsRefresh.Store(false)
	a.query = q
	a.results = nil
	a.selected = 0
	a.list.Position = layout.Position{}

	if q == "" {
		return
	}

	for _, p := range a.providers {
		a.results = append(a.results, p.Search(q)...)
	}
}

func (a *App) deleteWordBack() {
	selStart, selEnd := a.editor.Selection()
	if selStart > selEnd {
		selStart, selEnd = selEnd, selStart
	}
	text := a.editor.Text()
	runes := []rune(text)

	if selStart != selEnd {
		newRunes := make([]rune, 0, len(runes)-(selEnd-selStart))
		newRunes = append(newRunes, runes[:selStart]...)
		newRunes = append(newRunes, runes[selEnd:]...)
		a.editor.SetText(string(newRunes))
		a.editor.SetCaret(selStart, selStart)
		return
	}

	pos := selStart
	for pos > 0 && unicode.IsSpace(runes[pos-1]) {
		pos--
	}
	for pos > 0 && !unicode.IsSpace(runes[pos-1]) {
		pos--
	}

	if pos < selStart {
		newRunes := make([]rune, 0, len(runes)-(selStart-pos))
		newRunes = append(newRunes, runes[:pos]...)
		newRunes = append(newRunes, runes[selStart:]...)
		a.editor.SetText(string(newRunes))
		a.editor.SetCaret(pos, pos)
	}
}

func (a *App) moveWordLeft() {
	selStart, selEnd := a.editor.Selection()
	pos := min(selStart, selEnd)
	text := a.editor.Text()
	runes := []rune(text)

	for pos > 0 && unicode.IsSpace(runes[pos-1]) {
		pos--
	}
	for pos > 0 && !unicode.IsSpace(runes[pos-1]) {
		pos--
	}

	a.editor.SetCaret(pos, pos)
}

func (a *App) moveWordRight() {
	selStart, selEnd := a.editor.Selection()
	pos := max(selStart, selEnd)
	text := a.editor.Text()
	runes := []rune(text)

	for pos < len(runes) && unicode.IsSpace(runes[pos]) {
		pos++
	}
	for pos < len(runes) && !unicode.IsSpace(runes[pos]) {
		pos++
	}

	a.editor.SetCaret(pos, pos)
}

func (a *App) layout(gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, a.colors.background)

	size := gtx.Constraints.Max
	radius := gtx.Dp(unit.Dp(4))
	r := clip.UniformRRect(image.Rect(0, 0, size.X, size.Y), radius)
	borderWidth := float32(2 * gtx.Dp(unit.Dp(1)))
	stroke := clip.Stroke{
		Path:  r.Path(gtx.Ops),
		Width: borderWidth,
	}.Op()
	paint.FillShape(gtx.Ops, a.colors.divider, stroke)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutInput(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutDivider(gtx)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return a.layoutResults(gtx)
		}),
	)
}

func (a *App) layoutInput(gtx layout.Context) layout.Dimensions {
	h := gtx.Dp(unit.Dp(inputHeightDp))
	gtx.Constraints.Min.Y = h
	gtx.Constraints.Max.Y = h
	padding := layout.UniformInset(unit.Dp(12))
	return padding.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		ed := material.Editor(a.theme, &a.editor, "Search applications...")
		ed.TextSize = unit.Sp(18)
		ed.Color = a.colors.primary
		ed.HintColor = color.NRGBA{R: 140, G: 140, B: 160, A: 200}
		ed.SelectionColor = color.NRGBA{R: 100, G: 140, B: 220, A: 100}
		return ed.Layout(gtx)
	})
}

func (a *App) layoutDivider(gtx layout.Context) layout.Dimensions {
	height := gtx.Dp(unit.Dp(1))
	w := gtx.Constraints.Max.X
	r := clip.Rect{Max: image.Pt(w, height)}
	paint.FillShape(gtx.Ops, a.colors.divider, r.Op())
	return layout.Dimensions{Size: image.Pt(w, height)}
}

func (a *App) layoutResults(gtx layout.Context) layout.Dimensions {
	if len(a.results) == 0 {
		return layout.Dimensions{}
	}

	a.list.Axis = layout.Vertical
	return a.list.Layout(gtx, len(a.results), func(gtx layout.Context, index int) layout.Dimensions {
		return a.layoutResult(gtx, index)
	})
}

func (a *App) layoutResult(gtx layout.Context, index int) layout.Dimensions {
	r := a.results[index]
	selected := index == a.selected

	height := gtx.Dp(unit.Dp(resultHeightDp))
	gtx.Constraints.Min.Y = height
	gtx.Constraints.Max.Y = height

	padding := layout.Inset{Top: unit.Dp(4), Bottom: unit.Dp(4), Left: unit.Dp(12), Right: unit.Dp(12)}

	macro := op.Record(gtx.Ops)
	dims := padding.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.layoutIcon(gtx, r.Icon)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return a.layoutResultText(gtx, r, selected)
			}),
		)
	})
	call := macro.Stop()

	dims.Size.Y = height

	if selected {
		paint.FillShape(gtx.Ops,
			a.colors.selection,
			clip.Rect{Max: dims.Size}.Op(),
		)
	}

	call.Add(gtx.Ops)
	return dims
}

func (a *App) layoutIcon(gtx layout.Context, img image.Image) layout.Dimensions {
	if img == nil {
		img = placeholderIcon()
	}

	size := gtx.Dp(unit.Dp(32))
	gtx.Constraints = layout.Exact(image.Pt(size, size))

	img = scaleImage(img, size)
	imgOp := paint.NewImageOp(img)
	imgOp.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: image.Pt(size, size)}
}

func (a *App) layoutResultText(gtx layout.Context, r search.Result, selected bool) layout.Dimensions {
	nameColor := a.colors.primary
	descColor := a.colors.secondary
	if selected {
		nameColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		descColor = color.NRGBA{R: 200, G: 200, B: 220, A: 255}
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(a.theme, r.Name)
			label.Color = nameColor
			label.TextSize = unit.Sp(15)
			label.MaxLines = 1
			return label.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if r.Description == "" {
				return layout.Dimensions{}
			}
			label := material.Body2(a.theme, r.Description)
			label.Color = descColor
			label.TextSize = unit.Sp(12)
			label.MaxLines = 1
			return label.Layout(gtx)
		}),
	)
}

func scaleImage(src image.Image, size int) image.Image {
	sb := src.Bounds()
	if sb.Dx() == size && sb.Dy() == size {
		return src
	}
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, sb, xdraw.Over, nil)
	return dst
}

func placeholderIcon() image.Image {
	const s = 32
	img := image.NewRGBA(image.Rect(0, 0, s, s))
	bg := color.NRGBA{R: 80, G: 80, B: 100, A: 255}
	for y := range s {
		for x := range s {
			img.Set(x, y, bg)
		}
	}
	return img
}

func newTheme(colors themeConfig) *material.Theme {
	th := material.NewTheme()
	th.Bg = colors.background
	th.Fg = colors.primary
	th.ContrastBg = color.NRGBA{R: 137, G: 180, B: 250, A: 255}
	th.ContrastFg = color.NRGBA{R: 30, G: 30, B: 46, A: 255}
	if colors.typeface != "" {
		th.Face = colors.typeface
	}
	return th
}

func parseThemeColors(c config.ThemeConfig) themeConfig {
	return themeConfig{
		background: mustParseColor(c.Background),
		divider:    mustParseColor(c.Divider),
		primary:    mustParseColor(c.Primary),
		secondary:  mustParseColor(c.Secondary),
		selection:  mustParseColor(c.Selection),
		typeface:   font.Typeface(c.Font),
	}
}

func mustParseColor(s string) color.NRGBA {
	c, err := parseHexColor(s)
	if err != nil {
		log.Fatalf("invalid colour %q: %v", s, err)
	}
	return c
}

func parseHexColor(s string) (color.NRGBA, error) {
	s = strings.TrimPrefix(s, "#")

	var r, g, b, a uint8
	a = 255

	switch len(s) {
	case 6:
		v, err := strconv.ParseUint(s, 16, 24)
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("invalid hex colour: %w", err)
		}
		r = uint8(v >> 16)
		g = uint8(v >> 8)
		b = uint8(v)
	case 8:
		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("invalid hex colour: %w", err)
		}
		r = uint8(v >> 24)
		g = uint8(v >> 16)
		b = uint8(v >> 8)
		a = uint8(v)
	default:
		return color.NRGBA{}, fmt.Errorf("colour must be #RRGGBB or #RRGGBBAA, got %d hex digits", len(s))
	}

	return color.NRGBA{R: r, G: g, B: b, A: a}, nil
}
