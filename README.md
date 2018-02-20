# plexStats

plexStats is a project I did at my university ([HTWK Leipzig](https://www.htwk-leipzig.de/startseite/))
to get my accreditation for the exam of the databases module.

It provides a webservice with [Gin](https://github.com/gin-gonic/gin) with a
[webhook endpoint](https://support.plex.tv/articles/115002267687-webhooks/) for Plex
and gathers all received information in a SQLite database. The frontend shows basic
usage statistics like total plays by hour.

![example image of the frontend](https://i.imgur.com/PfC78Fy.png)

I got inspired by [Tautulli](https://github.com/Tautulli/Tautulli) which provides
full stack monitoring and tracking for Plex Media Servers.

# Deployment

Thanks to [go-bindata](https://github.com/tmthrgd/go-bindata
) you can deploy it with a single binary from the [releases page](https://github.com/hashworks/plexStats/releases).
Just pick one for your architecture and launch it, it will provide you with a web server
running on http://localhost:65431/. The webhook is available on http://localhost:65431/webhook.

Note that it will create a SQLite Database in a file named `plex.db` in your current working directory.

## By source

If you want to build by source you have to install the following dependencies:
* [Golang](https://golang.org/doc/install)
* [sassc](https://github.com/sass/sassc)
* [npm](https://www.npmjs.com/get-npm)

Afterwards you can install the Go dependencies.
```sh
make dependencies
```

And build plexStats which will provide you with a binary under `./bin/plexStats`.

```sh
make build
```
