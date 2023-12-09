package routes

import (
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/rharshit82/url-shortner/database"
)

func ResolveUrl(c *fiber.Ctx) error {
	url := c.Params("url")

	r := database.CreateClient(0)
	defer r.Close()

	value, err := r.Get(database.Ctx, url).Result()
	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Short not found in the database",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Can not connect to DB",
		})
	}

	rInr := database.CreateClient(1)

	defer rInr.Close()

	_ = rInr.Incr(database.Ctx, "counter")
	return c.Redirect(value, 301)

}
