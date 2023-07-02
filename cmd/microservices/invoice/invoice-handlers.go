package main

import (
	"fmt"
	"github.com/phpdave11/gofpdf"
	"github.com/phpdave11/gofpdf/contrib/gofpdi"
	"net/http"
	"time"
)

type Order struct {
	ID        int       `json:"id"`
	Quantity  int       `json:"quantity"`
	Amount    int       `json:"amount"`
	Product   string    `json:"product"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func (application *application) handlerPostCreateAndSendInvoice(w http.ResponseWriter, r *http.Request) {
	// receive json
	var order Order
	application.readJSONInto(&order, w, r)

	// Mock
	// order.ID = 80
	// order.Email = "test@test.com"
	// order.Amount = 1000
	// order.Quantity = 1
	// order.Product = "Widget"
	// order.FirstName = "Name"
	// order.LastName = "Surname"
	// order.CreatedAt = time.Now()
	// Mock end

	// generate pdf with invoice data
	application.getInvoicePDF(order)

	// create mail
	attachments := []string{fmt.Sprintf("./invoices/%d.pdf", order.ID)}

	// send mail with attachment
	err := application.SendEmail("info@widget.com", order.Email, "Widget Invoice", "invoice", attachments, nil)
	if err != nil {
		application.errorLog.Println("cannot send email with invoice ", err)
		return
	}

	// send response
	var answerPayload AnswerPayload
	answerPayload.Error = false
	answerPayload.Message = fmt.Sprintf("Invoice %d.pdf created and sent to %s", order.ID, order.Email)
	application.convertToJsonAndSend(answerPayload, w)
}

func (application *application) getInvoicePDF(order Order) {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(10, 13, 10)
	pdf.SetAutoPageBreak(true, 0)

	importer := gofpdi.NewImporter()
	template := importer.ImportPage(pdf, "./pdf-templates/invoice.pdf", 1, "/MediaBox")

	pdf.AddPage()
	importer.UseImportedTemplate(pdf, template, 0, 0, 215.9, 0)

	pdf.SetY(50)
	pdf.SetX(10)
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(
		97,
		8,
		fmt.Sprintf("Attention: %s %s", order.FirstName, order.LastName),
		"",
		0,
		"L",
		false,
		0,
		"",
	)
	pdf.Ln(5)
	pdf.CellFormat(97, 8, order.Email, "", 0, "L", false, 0, "")

	pdf.Ln(5)
	pdf.CellFormat(97, 8, order.CreatedAt.Format("2006-01-02"), "", 0, "L", false, 0, "")

	pdf.SetX(58)
	pdf.SetY(93)
	pdf.CellFormat(155, 8, order.Product, "", 0, "L", false, 0, "")

	pdf.SetX(166)
	pdf.CellFormat(20, 8, fmt.Sprintf("%d", order.Quantity), "", 0, "C", false, 0, "")

	pdf.SetX(185)
	pdf.CellFormat(20, 8, fmt.Sprintf("$%.2f", float64(order.Amount/100.0)), "", 0, "R", false, 0, "")

	invoicePath := fmt.Sprintf("./invoices/%d.pdf", order.ID)
	err := pdf.OutputFileAndClose(invoicePath)
	if err != nil {
		application.errorLog.Println("cannot output file and close ", err)
		return
	}
}
