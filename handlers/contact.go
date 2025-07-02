package handlers

import (
	"log"
	"net/http"
	"strings"
)

func HandleContact(w http.ResponseWriter, r *http.Request) {
	data := GetPageData(r)
	data.OfficeLocation = "123 Forum Street, Internet City"
	data.BusinessHours = "Monday-Friday: 9AM-5PM"

	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			RenderErrorPage(w, r, http.StatusBadRequest, "Error parsing form")
			return
		}

		data.FormData = map[string]string{
			"name":    r.FormValue("name"),
			"email":   r.FormValue("email"),
			"subject": r.FormValue("subject"),
			"message": r.FormValue("message"),
		}

		errors := make(map[string]string)

		if data.FormData["name"] == "" {
			errors["name"] = "Name is required"
		}

		if data.FormData["email"] == "" {
			errors["email"] = "Email is required"
		} else if !strings.Contains(data.FormData["email"], "@") {
			errors["email"] = "Invalid email format"
		}

		if data.FormData["subject"] == "" {
			errors["subject"] = "Subject is required"
		}

		if data.FormData["message"] == "" {
			errors["message"] = "Message is required"
		} else if len(data.FormData["message"]) < 10 {
			errors["message"] = "Message should be at least 10 characters"
		}

		if len(errors) > 0 {
			data.Errors = errors
		} else {
			log.Printf("New contact message from %s (%s): %s - %s",
				data.FormData["name"],
				data.FormData["email"],
				data.FormData["subject"],
				data.FormData["message"])

			data.Success = true
			data.Message = "Thank you for your message! We'll get back to you soon."
			data.FormData = make(map[string]string)
		}
	}

	RenderTemplate(w, "contact", data)
}
