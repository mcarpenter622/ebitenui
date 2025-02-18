package widget

import (
	"image"
	"testing"

	"github.com/ebitenui/ebitenui/input"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/matryer/is"
	"github.com/stretchr/testify/mock"
)

type controlMock struct {
	mock.Mock
}

func TestContainer_Render(t *testing.T) {
	w := NewWidget()
	m := controlMock{}
	m.On("GetWidget").Maybe().Return(w)
	m.On("PreferredSize").Maybe().Return(50, 50)
	m.On("SetLocation", mock.Anything).Maybe()
	m.On("Render", mock.Anything, mock.Anything)

	c := newContainer(t,
		ContainerOpts.Layout(newRowLayout(t)))
	c.AddChild(&m)

	render(c, t)

	m.AssertExpectations(t)
}

func TestContainer_Render_AutoDisableChildren(t *testing.T) {
	is := is.New(t)

	w := NewWidget()
	m := controlMock{}
	m.On("GetWidget").Maybe().Return(w)
	m.On("PreferredSize").Maybe().Return(50, 50)
	m.On("SetLocation", mock.Anything).Maybe()
	m.On("Render", mock.Anything, mock.Anything).Maybe()

	c := newContainer(t,
		ContainerOpts.AutoDisableChildren(),
		ContainerOpts.Layout(newRowLayout(t)))
	c.AddChild(&m)

	c.widget.Disabled = true
	render(c, t)

	is.True(w.Disabled)
}

func TestContainer_SetupInputLayer(t *testing.T) {
	def := func(s input.SetupInputLayerFunc) {
		// nothing to do
	}

	w := NewWidget()
	m := controlMock{}
	m.On("GetWidget").Maybe().Return(w)
	m.On("SetupInputLayer", mock.AnythingOfType("input.DeferredSetupInputLayerFunc"))

	c := newContainer(t,
		ContainerOpts.Layout(newRowLayout(t)))
	c.AddChild(&m)

	c.SetupInputLayer(def)

	m.AssertExpectations(t)
}

func TestContainer_MinMaxWidthHeight(t *testing.T) {
	is := is.New(t)
	w := NewWidget()
	is.Equal(w.MinWidth, 0)
	is.Equal(w.MinHeight, 0)
	is.Equal(w.MaxWidth, 0)
	is.Equal(w.MaxHeight, 0)
	m := controlMock{}
	m.On("GetWidget").Maybe().Return(w)
	m.On("PreferredSize").Maybe().Return(50, 50)
	m.On("SetLocation", mock.Anything).Maybe()
	m.On("Render", mock.Anything, mock.Anything)

	c := newContainer(t,
		ContainerOpts.WidgetOpts(
			WidgetOpts.MinSize(100, 100),
		))
	c.AddChild(&m)

	width, height := c.PreferredSize()
	is.Equal(width, 100)
	is.Equal(height, 100)

	// setting max size should not affect preferred size
	c = newContainer(t,
		ContainerOpts.WidgetOpts(
			WidgetOpts.MaxSize(200, 200),
		))
	c.AddChild(&m)

	width, height = c.PreferredSize()
	is.Equal(width, 0)
	is.Equal(height, 0)

	render(c, t)
	m.AssertExpectations(t)
}

func (c *controlMock) GetWidget() *Widget {
	args := c.Called()
	if arg, ok := args.Get(0).(*Widget); ok {
		return arg
	}
	return nil
}

func (c *controlMock) PreferredSize() (int, int) {
	c.Called()
	return -1, -1
}

func (c *controlMock) SetLocation(rect image.Rectangle) {
	c.Called(rect)
}

func (c *controlMock) Render(screen *ebiten.Image) {
	c.Called(screen)
}

func (c *controlMock) SetupInputLayer(def input.DeferredSetupInputLayerFunc) {
	c.Called(def)
}

func newContainer(t *testing.T, opts ...ContainerOpt) *Container {
	t.Helper()
	return NewContainer(opts...)
}
