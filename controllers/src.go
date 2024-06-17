package controllers

import "regexp"

func HtmlSrcReplacer(html string) string {
	regex := regexp.MustCompile(`src=["']([^"']+)["']`)
	return regex.ReplaceAllString(html, `src="https://localhost:8443/staticresolver9F3DQN?url=$1"`)
}
