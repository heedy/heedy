
if (app.getUsername().length > 0) {
	app.render(<MainPage />);
} else {
	app.render(<LoginForm />);
}
