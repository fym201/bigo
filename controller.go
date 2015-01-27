package bigo

type Controller struct {
	App *Bigo
}

type IController interface {
	Init(app *Bigo)
}

func (c *Controller) Init(app *Bigo) {
	c.App = app
}
