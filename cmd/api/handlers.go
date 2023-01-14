package main

import (
	"net/http"

	"github.com/mofodox/anymail/internal/data"
	"github.com/mofodox/anymail/internal/request"
	"github.com/mofodox/anymail/internal/response"
)

// Status Handler
func (app *application) status(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"Status": "OK",
	}

	err := response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

// MyTengahSGMttp Handler
func (app *application) sendEmailMyTengahMTPP(w http.ResponseWriter, r *http.Request) {
	var input struct {
		MtppNumber string `json:"mtpp_number"`
		MtppUrl    string `json:"mtpp_url"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	mtppCard := &data.MTPPCard{
		MTPPNumber: input.MtppNumber,
		MTPPUrl:    input.MtppUrl,
	}

	err = app.models.MTPPCards.Insert(mtppCard)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = response.JSON(w, http.StatusCreated, mtppCard)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	emailData := data.MTPPCard{
		MTPPNumber: mtppCard.MTPPNumber,
		MTPPUrl:    mtppCard.MTPPUrl,
	}

	err = app.mailer.Send("khairulakmal.dev@gmail.com", emailData, "mtpp_mail.tmpl")
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
