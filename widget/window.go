package widget

import (
	"image"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/input"
	"github.com/hajimehoshi/ebiten/v2"
)

type RemoveWindowFunc func()

type WindowChangedEventArgs struct {
	Window *Window
	Rect   image.Rectangle
}

type WindowChangedHandlerFunc func(args *WindowChangedEventArgs)

type Window struct {
	ResizeEvent *event.Event
	MoveEvent   *event.Event

	Modal      bool
	Contents   *Container
	TitleBar   *Container
	Draggable  bool
	Resizeable bool
	MinSize    *image.Point
	MaxSize    *image.Point

	closeOnClick bool
	closeFunc    RemoveWindowFunc
	container    *Container

	titleBarHeight int

	startingPoint  image.Point
	dragging       bool
	resizing       bool
	resizingWidth  bool
	resizingHeight bool
	originalSize   image.Point
	init           *MultiOnce
}

type WindowOpt func(w *Window)

type WindowOptions struct {
}

var WindowOpts WindowOptions

func NewWindow(opts ...WindowOpt) *Window {
	w := &Window{
		MoveEvent:   &event.Event{},
		ResizeEvent: &event.Event{},
		init:        &MultiOnce{},
	}

	for _, o := range opts {
		o(w)
	}

	if w.TitleBar != nil {
		w.container = NewContainer(ContainerOpts.Layout(NewGridLayout(
			GridLayoutOpts.Columns(1),
			GridLayoutOpts.Stretch([]bool{true}, []bool{false, true}),
		)))
		w.TitleBar.GetWidget().LayoutData = GridLayoutData{MaxHeight: w.titleBarHeight}
		w.TitleBar.GetWidget().MinHeight = w.titleBarHeight
		if w.Draggable {
			w.TitleBar.GetWidget().MouseButtonPressedEvent.AddHandler(func(_ interface{}) {
				x, y := input.CursorPosition()
				w.startingPoint = image.Point{x, y}
				w.dragging = true
			})
			w.TitleBar.GetWidget().MouseButtonReleasedEvent.AddHandler(func(_ interface{}) {
				if w.dragging {
					w.dragging = false
					w.MoveEvent.Fire(&WindowChangedEventArgs{
						Window: w,
						Rect:   w.container.GetWidget().Rect,
					})
				}
			})
		}
		w.container.AddChild(w.TitleBar)
		w.container.AddChild(w.Contents)
	} else {
		w.container = NewContainer(ContainerOpts.Layout(NewGridLayout(
			GridLayoutOpts.Columns(1),
			GridLayoutOpts.Stretch([]bool{true}, []bool{true}),
		)))

		w.container.AddChild(w.Contents)
	}

	if w.Resizeable {
		w.Contents.GetWidget().MouseButtonPressedEvent.AddHandler(func(_ interface{}) {
			x, y := input.CursorPosition()
			w.startingPoint = image.Point{x, y}
			w.originalSize.X = w.container.GetWidget().Rect.Max.X
			w.originalSize.Y = w.container.GetWidget().Rect.Max.Y
			w.resizing = true
		})
		w.Contents.GetWidget().MouseButtonReleasedEvent.AddHandler(func(_ interface{}) {
			if w.resizing {
				w.resizing = false
				w.ResizeEvent.Fire(&WindowChangedEventArgs{
					Window: w,
					Rect:   w.container.GetWidget().Rect,
				})
			}
		})
	}

	if w.closeOnClick {
		w.container.GetWidget().MouseButtonReleasedEvent.AddHandler(func(args interface{}) {
			a := args.(*WidgetMouseButtonReleasedEventArgs)
			if !a.Inside && w.closeFunc != nil {
				w.closeFunc()
			}
		})
	}

	w.init.Do()
	return w
}

// This is the container with the body of this window
func (o WindowOptions) Contents(c *Container) WindowOpt {
	return func(w *Window) {
		w.Contents = c
	}
}

// Sets the container for the TitleBar and its fixed height
func (o WindowOptions) TitleBar(tb *Container, height int) WindowOpt {
	return func(w *Window) {
		w.TitleBar = tb
		w.titleBarHeight = height
	}
}

// Sets the window to be modal. Blocking UI interactions on anything else.
func (o WindowOptions) Modal() WindowOpt {
	return func(w *Window) {
		w.Modal = true
	}
}

// Sets the window to be draggable. The handle for this is the titleBar.
// If you haven't provided a titleBar this option is ignored
func (o WindowOptions) Draggable() WindowOpt {
	return func(w *Window) {
		w.Draggable = true
	}
}

// Sets the window to be resizeable
func (o WindowOptions) Resizeable() WindowOpt {
	return func(w *Window) {
		w.Resizeable = true
	}
}

// Sets the minimum size that the window can be reszied to
func (o WindowOptions) MinSize(width int, height int) WindowOpt {
	return func(w *Window) {
		w.MinSize = &image.Point{X: width, Y: height}
	}
}

// Set the maximum size that the window can be resized to
func (o WindowOptions) MaxSize(width int, height int) WindowOpt {
	return func(w *Window) {
		w.MaxSize = &image.Point{X: width, Y: height}
	}
}

// Close the window if the user clicks outside of the window
func (o WindowOptions) CloseOnClickOut() WindowOpt {
	return func(w *Window) {
		w.closeOnClick = true
	}
}

// This handler is triggered when a move event is completed
func (o WindowOptions) MoveHandler(f WindowChangedHandlerFunc) WindowOpt {
	return func(w *Window) {
		w.MoveEvent.AddHandler(func(args interface{}) {
			f(args.(*WindowChangedEventArgs))
		})
	}
}

// This handler is triggered when a resize event is completed
func (o WindowOptions) ResizeHandler(f WindowChangedHandlerFunc) WindowOpt {
	return func(w *Window) {
		w.ResizeEvent.AddHandler(func(args interface{}) {
			f(args.(*WindowChangedEventArgs))
		})
	}
}

// This option sets the size and location of the window.
// This method will account for specified MinSize and MaxSize values.
func (o WindowOptions) Location(rect image.Rectangle) WindowOpt {
	return func(w *Window) {
		w.init.Append(func() { w.container.SetLocation(rect) })
	}
}

// This method is used to be able to close the window
func (w *Window) Close() {
	if w.closeFunc != nil {
		w.closeFunc()
	}
}

// This method will set the size and location of this window.
// This method will account for specified MinSize and MaxSize values.
func (w *Window) SetLocation(rect image.Rectangle) {
	if w.MinSize != nil {
		if rect.Dx() < w.MinSize.X {
			rect.Max.X = rect.Min.X + w.MinSize.X
		}
		if rect.Dy() < w.MinSize.Y {
			rect.Max.Y = rect.Min.Y + w.MinSize.Y
		}
	}

	if w.MaxSize != nil {
		if rect.Dx() > w.MaxSize.X {
			rect.Max.X = rect.Min.X + w.MaxSize.X
		}
		if rect.Dy() > w.MaxSize.Y {
			rect.Max.Y = rect.Min.Y + w.MaxSize.Y
		}
	}

	w.container.SetLocation(rect)
}

// Typically used internally.
func (w *Window) SetCloseFunction(close RemoveWindowFunc) {
	w.closeFunc = close
}

// Typically used internally
func (w *Window) RequestRelayout() {
	w.container.RequestRelayout()
}

// Typically used internally
func (w *Window) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	w.container.GetWidget().ElevateToNewInputLayer(&input.Layer{
		DebugLabel: "window",
		EventTypes: input.LayerEventTypeAll,
		BlockLower: true,
		FullScreen: w.Modal,
		RectFunc: func() image.Rectangle {
			return w.container.GetWidget().Rect
		},
	})
}

// Typically used internally
func (w *Window) Render(screen *ebiten.Image, def DeferredRenderFunc) {
	x, y := input.CursorPosition()

	if w.dragging {
		if w.startingPoint.X != x || w.startingPoint.Y != y {
			newRect := w.container.GetWidget().Rect.Add(image.Point{x - w.startingPoint.X, y - w.startingPoint.Y})
			w.SetLocation(newRect)
			w.RequestRelayout()
			w.startingPoint = image.Point{x, y}
		}
	}
	if w.resizing {
		if w.startingPoint.X != x || w.startingPoint.Y != y {
			if w.resizingWidth {
				newRect := w.container.GetWidget().Rect
				newRect.Max.X = w.originalSize.X - (w.startingPoint.X - x)
				w.SetLocation(newRect)
				w.RequestRelayout()
			}
			if w.resizingHeight {
				newRect := w.container.GetWidget().Rect
				newRect.Max.Y = w.originalSize.Y - (w.startingPoint.Y - y)

				w.SetLocation(newRect)
				w.RequestRelayout()
			}
		}
	}
	if w.Resizeable {
		if w.container.GetWidget().inputLayer.ActiveFor(x, y, input.LayerEventTypeAll) {
			xRect := image.Rect(w.container.GetWidget().Rect.Max.X-6, w.container.GetWidget().Rect.Min.Y, w.container.GetWidget().Rect.Max.X, w.container.GetWidget().Rect.Max.Y)
			yRect := image.Rect(w.container.GetWidget().Rect.Min.X, w.container.GetWidget().Rect.Max.Y-6, w.container.GetWidget().Rect.Max.X, w.container.GetWidget().Rect.Max.Y)
			cursorRect := image.Rect(x, y, x+1, y+1)
			if cursorRect.Overlaps(xRect) {
				ebiten.SetCursorShape(ebiten.CursorShapeEWResize)
				w.resizingWidth = true
				w.resizingHeight = false
			} else if cursorRect.Overlaps(yRect) {
				ebiten.SetCursorShape(ebiten.CursorShapeNSResize)
				w.resizingWidth = false
				w.resizingHeight = true
			} else {
				ebiten.SetCursorShape(ebiten.CursorShapeDefault)
				w.resizingWidth = false
				w.resizingHeight = false
			}
		} else {
			ebiten.SetCursorShape(ebiten.CursorShapeDefault)
			w.resizingWidth = false
			w.resizingHeight = false
		}
	}
	w.container.Render(screen, def)
}
