## Trendever backend services


# Fast local deploy instruction

* Install go and [gb](https://getgb.io/).
* Clone this repository and run `gb build`.
* Install docker and docker-compose; enter local/ directory; launch `docker-compose up` then `docker-compose start`. Everything should be running now.

# Default configs

Default local-deploy configs are located in `configs/`. Important open ports:

* 8080: api
* 8087: mandible
* 5432: postgres (db container)
* 3004: core (visit localhost:3004/qor for web admin interface)
