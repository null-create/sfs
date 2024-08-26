package client

import (
	"html/template"
	"log"
	"path/filepath"
)

func GetTemplatePaths() map[string]string {
	var templateFolder = "./static/templates"
	var templates = []string{
		templateFolder + "/index.html",
		templateFolder + "/file.html",
		templateFolder + "/folder.html",
		templateFolder + "/frame.html",
		templateFolder + "/toolbar.html",
		templateFolder + "/header.html",
		templateFolder + "/trash.html",
		templateFolder + "/recent.html",
		templateFolder + "/starred.html",
		templateFolder + "/error.html",
		templateFolder + "/table.html",
		templateFolder + "/add.html",
		templateFolder + "/user.html",
		templateFolder + "/upload.html",
		templateFolder + "/edit.html",
		templateFolder + "/settings.html",
		templateFolder + "/search.html",
	}
	var tmplPaths = make(map[string]string)
	for _, tpath := range templates {
		var tpathAbs, err = filepath.Abs(tpath)
		if err != nil {
			log.Fatal(err)
		}
		tmplPaths[filepath.Base(tpath)] = tpathAbs
	}
	return tmplPaths
}

func (c *Client) ParseTemplates() error {
	var tmplPaths = GetTemplatePaths()
	tmpl, err := template.ParseFiles(
		tmplPaths["index.html"],
		tmplPaths["file.html"],
		tmplPaths["folder.html"],
		tmplPaths["frame.html"],
		tmplPaths["toolbar.html"],
		tmplPaths["header.html"],
		tmplPaths["trash.html"],
		tmplPaths["recent.html"],
		tmplPaths["error.html"],
		tmplPaths["table.html"],
		tmplPaths["add.html"],
		tmplPaths["user.html"],
		tmplPaths["upload.html"],
		tmplPaths["edit.html"],
		tmplPaths["settings.html"],
		tmplPaths["search.html"],
	)
	if err != nil {
		return err
	}
	c.Templates = tmpl
	return nil
}
