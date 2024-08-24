package client

import (
	"html/template"
	"log"
	"path/filepath"
)

func GetTemplatePaths() map[string]string {
	var templates = []string{
		"./pkg/client/views/templates/index.html",
		"./pkg/client/views/templates/file.html",
		"./pkg/client/views/templates/folder.html",
		"./pkg/client/views/templates/frame.html",
		"./pkg/client/views/templates/toolbar.html",
		"./pkg/client/views/templates/header.html",
		"./pkg/client/views/templates/trash.html",
		"./pkg/client/views/templates/recent.html",
		"./pkg/client/views/templates/starred.html",
		"./pkg/client/views/templates/error.html",
		"./pkg/client/views/templates/table.html",
		"./pkg/client/views/templates/add.html",
		"./pkg/client/views/templates/user.html",
		"./pkg/client/views/templates/upload.html",
		"./pkg/client/views/templates/edit.html",
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
	)
	if err != nil {
		return err
	}
	c.Templates = tmpl
	return nil
}
