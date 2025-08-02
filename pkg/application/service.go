package application

type Service struct {
	*App
}

func (app *App) Service() Service {
	return Service{App: app}
}
