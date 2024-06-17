package controllers_test

import (
	"devilginx/controllers"
	"testing"
)

func TestHtmlSrcReplacer(t *testing.T) {
	testHTML := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test</title>
		</head>
		<body>
			<img src="http://staticasset.com/image.jpg" />
			<img src="https://staticasset.com/script.js" />
			<img src="https://staticasset.com/assets/style.css" />
		</body>	
		</html>
	`

	expected := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test</title>
		</head>
		<body>
			<img src="https://localhost:8443/staticresolver9F3DQN?url=http://staticasset.com/image.jpg" />
			<img src="https://localhost:8443/staticresolver9F3DQN?url=https://staticasset.com/script.js" />
			<img src="https://localhost:8443/staticresolver9F3DQN?url=https://staticasset.com/assets/style.css" />
		</body>
		</html>
	`

	result := controllers.HtmlSrcReplacer(testHTML)
	t.Logf("Result: %s", result)

	if result != expected {
		t.Errorf("Expected: %s\nGot: %s", expected, result)
	}
}
