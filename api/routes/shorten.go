package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rharshit82/url-shortner/database"
	"github.com/rharshit82/url-shortner/helpers"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenUrl(c *fiber.Ctx) error {
	body := new(request)

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Can not parse request json"})
	}

	// implement rate limiting

	r2 := database.CreateClient(1)

	defer r2.Close()
	var currentUserIP string = c.IP()
	currentUserIPCount, err := r2.Get(database.Ctx, currentUserIP).Result()
	currentUserIPCountInt, _ := strconv.Atoi(currentUserIPCount)
	if err == redis.Nil {
		_ = r2.Set(database.Ctx, currentUserIP, os.Getenv("API_QUOTA"), 50*60*time.Second).Err()
	} else {
		r2.Get(database.Ctx, currentUserIP).Result()
		if currentUserIPCountInt <= 0 {
			limit, _ := r2.TTL(database.Ctx, currentUserIP).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":           "Rate limit exceeded",
				"rate_limit_rest": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	// check if the input if an actual url
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}
	// check for domain error
	if !helpers.CheckDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Bad Domain in request"})
	}
	// enforce https, SSL
	body.URL = helpers.EnforceHTTP(body.URL)

	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)
	defer r.Close()

	existingUrlShortForCurrentID, _ := r.Get(database.Ctx, id).Result()

	if existingUrlShortForCurrentID != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "URL custom short is already in use",
		})
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = r.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to connect to server",
		})
	}

	var serverDomain string = os.Getenv("DOMAIN")
	var finalCustomShort string = serverDomain + "/" + id
	resp := response{
		URL:             body.URL,
		CustomShort:     finalCustomShort,
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}
	r2.Decr(database.Ctx, currentUserIP)

	currentUserIPCountFinal, _ := r2.Get(database.Ctx, currentUserIP).Result()

	resp.XRateRemaining, _ = strconv.Atoi(currentUserIPCountFinal)
	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()

	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute
	return c.Status(fiber.StatusOK).JSON(resp)
}
