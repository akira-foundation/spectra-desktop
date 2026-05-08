package laravel

import (
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

func init() {
	gofakeit.Seed(0)
}

func generateExample(name, fieldType string, rules []string) any {
	lowerName := strings.ToLower(name)
	_ = rules

	switch fieldType {
	case "email":
		return gofakeit.Email()
	case "uuid":
		return gofakeit.UUID()
	case "url":
		return gofakeit.URL()
	case "integer":
		if strings.Contains(lowerName, "id") {
			return gofakeit.Number(1, 9999)
		}
		return gofakeit.Number(0, 100)
	case "numeric":
		return gofakeit.Float64Range(0, 1000)
	case "boolean":
		return gofakeit.Bool()
	case "array":
		return []any{}
	case "object":
		return map[string]any{}
	case "date":
		return gofakeit.Date().Format("2006-01-02")
	case "file":
		return nil
	}

	switch {
	case strings.Contains(lowerName, "email"):
		return gofakeit.Email()
	case lowerName == "name", strings.HasSuffix(lowerName, "_name"), strings.HasSuffix(lowerName, "name"):
		if strings.Contains(lowerName, "first") {
			return gofakeit.FirstName()
		}
		if strings.Contains(lowerName, "last") {
			return gofakeit.LastName()
		}
		if strings.Contains(lowerName, "user") {
			return gofakeit.Username()
		}
		if strings.Contains(lowerName, "company") {
			return gofakeit.Company()
		}
		return gofakeit.Name()
	case strings.Contains(lowerName, "password"):
		return "password"
	case strings.Contains(lowerName, "phone"), strings.Contains(lowerName, "tel"):
		return gofakeit.Phone()
	case strings.Contains(lowerName, "url"), strings.Contains(lowerName, "website"), strings.Contains(lowerName, "site"):
		return gofakeit.URL()
	case strings.Contains(lowerName, "slug"):
		return gofakeit.Word() + "-" + gofakeit.Word()
	case strings.Contains(lowerName, "title"):
		return gofakeit.Sentence(4)
	case strings.Contains(lowerName, "description"), strings.Contains(lowerName, "body"), strings.Contains(lowerName, "content"), strings.Contains(lowerName, "comment"), strings.Contains(lowerName, "note"):
		return gofakeit.Paragraph(1, 2, 12, " ")
	case lowerName == "id", strings.HasSuffix(lowerName, "_id"):
		return gofakeit.Number(1, 9999)
	case strings.Contains(lowerName, "address"):
		return gofakeit.Address().Address
	case strings.Contains(lowerName, "city"):
		return gofakeit.City()
	case strings.Contains(lowerName, "country"):
		return gofakeit.Country()
	case strings.Contains(lowerName, "state"), strings.Contains(lowerName, "province"):
		return gofakeit.State()
	case strings.Contains(lowerName, "zip"), strings.Contains(lowerName, "postal"):
		return gofakeit.Zip()
	case strings.Contains(lowerName, "lat"):
		return gofakeit.Latitude()
	case strings.Contains(lowerName, "lon"), strings.Contains(lowerName, "lng"):
		return gofakeit.Longitude()
	case strings.Contains(lowerName, "color"):
		return gofakeit.HexColor()
	case strings.Contains(lowerName, "ip"):
		return gofakeit.IPv4Address()
	case strings.Contains(lowerName, "currency"):
		return gofakeit.CurrencyShort()
	case strings.Contains(lowerName, "price"), strings.Contains(lowerName, "amount"), strings.Contains(lowerName, "cost"), strings.Contains(lowerName, "value"):
		return gofakeit.Float64Range(1, 1000)
	case strings.Contains(lowerName, "code"):
		return gofakeit.LetterN(6)
	case strings.Contains(lowerName, "token"):
		return gofakeit.LetterN(32)
	case strings.Contains(lowerName, "username"), strings.Contains(lowerName, "user_name"), strings.Contains(lowerName, "login"), strings.Contains(lowerName, "handle"):
		return gofakeit.Username()
	case strings.Contains(lowerName, "company"), strings.Contains(lowerName, "organization"):
		return gofakeit.Company()
	case strings.Contains(lowerName, "image"), strings.Contains(lowerName, "photo"), strings.Contains(lowerName, "avatar"):
		return "https://picsum.photos/640/480"
	case strings.Contains(lowerName, "date"), strings.Contains(lowerName, "_at"):
		return gofakeit.Date().Format(time.RFC3339)
	}

	return gofakeit.Word()
}

func buildExampleBody(fields []InferredField) map[string]any {
	body := make(map[string]any, len(fields))
	for _, f := range fields {
		body[f.Name] = f.Example
	}
	return body
}
