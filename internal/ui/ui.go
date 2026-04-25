package ui

import (
	"image"
	"image/color"
	"log"
	"os"
	"unicode"

	"gioui.org/app"
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

	"chameth.com/glauncher/internal/search"
)

type App struct {
	providers []search.Provider
	window    *app.Window
	editor    widget.Editor
	results   []search.Result
	selected  int
	query     string
	focused   bool
	navTag    struct{}
	theme     *material.Theme
}

func New(providers ...search.Provider) *App {
	a := &App{
		providers: providers,
		theme:     newTheme(),
	}
	a.editor.SingleLine = true
	return a
}

func (a *App) Run() {
	go func() {
		a.window = &app.Window{}
		a.window.Option(
			app.Size(unit.Dp(600), unit.Dp(420)),
			app.Decorated(false),
			app.Title("glauncher"),
		)
		if err := a.loop(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func (a *App) loop() error {
	var ops op.Ops

	for {
		switch e := a.window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			a.frame(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (a *App) frame(gtx layout.Context) {
	if !a.focused {
		gtx.Execute(key.FocusCmd{Tag: &a.editor})
		a.focused = true
	}

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
			}
		case key.NameDownArrow:
			if a.selected < len(a.results)-1 {
				a.selected++
			}
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
	if q == a.query {
		return
	}
	a.query = q
	a.results = nil
	a.selected = 0

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
	pos := selStart
	if selStart > selEnd {
		pos = selEnd
	}
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
	pos := selEnd
	if selStart > selEnd {
		pos = selStart
	}
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
	paint.Fill(gtx.Ops, a.theme.Bg)

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
	padding := layout.UniformInset(unit.Dp(16))
	return padding.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		ed := material.Editor(a.theme, &a.editor, "Search applications...")
		ed.TextSize = unit.Sp(18)
		ed.Color = a.theme.Fg
		ed.HintColor = color.NRGBA{R: 140, G: 140, B: 160, A: 200}
		ed.SelectionColor = color.NRGBA{R: 100, G: 140, B: 220, A: 100}
		return ed.Layout(gtx)
	})
}

func (a *App) layoutDivider(gtx layout.Context) layout.Dimensions {
	height := gtx.Dp(unit.Dp(1))
	w := gtx.Constraints.Max.X
	r := clip.Rect{Max: image.Pt(w, height)}
	paint.FillShape(gtx.Ops, color.NRGBA{R: 60, G: 60, B: 80, A: 255}, r.Op())
	return layout.Dimensions{Size: image.Pt(w, height)}
}

func (a *App) layoutResults(gtx layout.Context) layout.Dimensions {
	if len(a.results) == 0 {
		return layout.Dimensions{}
	}

	var dims layout.Dimensions
	list := &layout.List{Axis: layout.Vertical}
	dims = list.Layout(gtx, len(a.results), func(gtx layout.Context, index int) layout.Dimensions {
		return a.layoutResult(gtx, index)
	})
	return dims
}

func (a *App) layoutResult(gtx layout.Context, index int) layout.Dimensions {
	r := a.results[index]
	selected := index == a.selected

	padding := layout.UniformInset(unit.Dp(8))

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

	if selected {
		paint.FillShape(gtx.Ops,
			color.NRGBA{R: 100, G: 150, B: 230, A: 200},
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
	nameColor := a.theme.Fg
	descColor := color.NRGBA{R: 160, G: 160, B: 180, A: 255}
	if selected {
		nameColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		descColor = color.NRGBA{R: 200, G: 200, B: 220, A: 255}
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(a.theme, r.Name)
			label.Color = nameColor
			label.TextSize = unit.Sp(15)
			return label.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if r.Description == "" {
				return layout.Dimensions{}
			}
			label := material.Body2(a.theme, r.Description)
			label.Color = descColor
			label.TextSize = unit.Sp(12)
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
	for y := 0; y < s; y++ {
		for x := 0; x < s; x++ {
			img.Set(x, y, bg)
		}
	}
	return img
}

func newTheme() *material.Theme {
	th := material.NewTheme()
	th.Bg = color.NRGBA{R: 30, G: 30, B: 46, A: 240}
	th.Fg = color.NRGBA{R: 205, G: 214, B: 244, A: 255}
	th.ContrastBg = color.NRGBA{R: 137, G: 180, B: 250, A: 255}
	th.ContrastFg = color.NRGBA{R: 30, G: 30, B: 46, A: 255}
	return th
}
