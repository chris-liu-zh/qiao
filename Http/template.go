/*
 * @Author: Chris
 * @Date: 2024-05-16 22:38:04
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-14 10:54:28
 * @Description: 请填写简介
 */
package Http

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"text/template"
)

var t *HtmlTemplate

type HtmlTemplate struct {
	template *template.Template
}

func NewTemplates(templateDir ...string) (err error) {
	t = &HtmlTemplate{
		template: template.New(""),
	}

	// 使用 filepath.Glob 查找所有匹配的 HTML 文件
	var files []string
	for _, dir := range templateDir {
		filesPath, err := filepath.Glob(dir)
		if err != nil {
			return err
		}
		files = append(files, filesPath...)
	}

	for _, path := range files {
		dir := filepath.Dir(path)
		base := filepath.Base(path)
		nameWithoutExt := base[:len(base)-len(filepath.Ext(base))]
		templateName := filepath.Join(dir, nameWithoutExt)
		slog.Info("加载模版", "名称", templateName)
		if _, err := t.template.New(templateName).ParseFiles(path); err != nil {
			return err
		}
	}
	return nil
}

func Html(w http.ResponseWriter, templateName string, data any) {
	t := t.template.Lookup(templateName)
	if err := t.Execute(w, data); err != nil {
		NotFound(w, " Template not found: "+templateName)
	}
}

// func (t *Tpl) Execute(w http.ResponseWriter, tplName string, data any) error {
// 	return t.T.ExecuteTemplate(w, tplName, data)
// }

// func (t *Template) htmlCache(tplName string, data any) (htmlName string, err error) {
// 	buf := new(bytes.Buffer)
// 	htmlName = t.HtmlCachePath + "/" + tplName
// 	if err = t.Tpl.ExecuteTemplate(buf, tplName, data); err != nil {
// 		return
// 	}
// 	if err = os.WriteFile(htmlName, buf.Bytes(), 0660); err != nil {
// 		return
// 	}
// 	return
// }

// func (t *Template) Execute(w http.ResponseWriter, tplName string, data any) {

// 	htmlFile := t.HtmlCachePath + "/" + tplName
// 	if t.HtmlCache {
// 		if CheckFileIsExist(htmlFile) {
// 			html, err := ReadFile(htmlFile)
// 			if err != nil {
// 				return
// 			}
// 			w.Write(html)
// 			return
// 		} else {
// 			go t.htmlCache(tplName, data)
// 		}
// 	}

// 	t.Tpl.ExecuteTemplate(w, tplName, data)
// }
