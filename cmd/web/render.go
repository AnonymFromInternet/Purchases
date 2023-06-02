package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type templateData struct {
	StringMap       map[string]string
	IntMap          map[string]int
	FloatMap        map[string]float32
	Data            map[string]interface{}
	CSRFToken       string
	Error           string
	IsAuthenticated bool
	Api             string
	CssVersion      string
}

var functions = template.FuncMap{
	"getConvertedPrice": getConvertedPrice,
}

func getConvertedPrice(notConvertedPrice int) string {
	return fmt.Sprintf("$%.2f", float64(notConvertedPrice/100))
}

//go:embed templates
var templateFS embed.FS

func (application *application) addDefaultData(data *templateData, r *http.Request) *templateData {
	data.Api = application.config.api
	return data
}

func (application *application) renderTemplate(w http.ResponseWriter, r *http.Request, pageName string, td *templateData, partials ...string) error {
	var tmpl *template.Template
	var err error

	templateToRender := fmt.Sprintf("templates/%s.page.gohtml", pageName)

	existedTemplate, isTemplateInMap := application.templateCache[templateToRender]
	if application.config.env == "production" && isTemplateInMap {
		tmpl = existedTemplate
	} else {
		tmpl, err = application.parseTemplate(partials, pageName, templateToRender)
		if err != nil {
			application.errorLog.Println(err)

			return err
		}
	}

	if td == nil {
		td = &templateData{}
	}

	td = application.addDefaultData(td, r)

	err = tmpl.Execute(w, td)
	if err != nil {
		application.errorLog.Println(err)

		return err
	}

	return nil
}

func (application *application) parseTemplate(partials []string, pageName string, templateToRender string) (*template.Template, error) {
	var tmpl *template.Template
	var err error

	// build partials
	if len(partials) > 0 {
		for index, value := range partials {
			partials[index] = fmt.Sprintf("templates/%s.partial.gohtml", value)
		}

		// Скорее всего тут в темплейт с именем "%s.page.gohtml", pageName, например main.page.gohtml парсятся данные из
		// base.layout.gohtml, partials и "templates/%s.page.gohtml", pageName, например templates/main.page.gohtml
		// То есть файлы как бы мерджаться в один файл с именем например main.page.gohtml
		tmpl, err = template.New(fmt.Sprintf("%s.page.gohtml", pageName)).Funcs(functions).ParseFS(
			templateFS,
			"templates/base.layout.gohtml",
			strings.Join(partials, ","),
			templateToRender,
		)
	} else {
		tmpl, err = template.New(fmt.Sprintf("%s.page.gohtml", pageName)).Funcs(functions).ParseFS(
			templateFS,
			"templates/base.layout.gohtml",
			templateToRender,
		)
	}

	application.templateCache[templateToRender] = tmpl

	return tmpl, err
}
