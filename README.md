# url-shortener

## Environment variables

Config is loaded either from a YAML file (`CONFIG_PATH` set, e.g. local `make run`) or
directly from env vars (`CONFIG_PATH` unset, e.g. Docker/Railway).

| Variable            | Required | Default | Description                                  |
|---------------------|----------|---------|-----------------------------------------------|
| `CONFIG_PATH`       | no       | —       | Path to YAML config; if unset, reads env vars below directly |
| `ENV`               | no       | `dev`   | Environment name (`dev`/`test`/`prod`), logging format |
| `STORAGE_PATH`      | yes      | —       | Path to the sqlite database file              |
| `HTTP_ADDRESS`      | yes      | —       | Address the server listens on (`host:port`)   |
| `HTTP_TIMEOUT`      | no       | `4s`    | Request read/write timeout                    |
| `HTTP_IDLE_TIMEOUT` | no       | `60s`   | Keep-alive idle timeout                       |
| `HTTP_USER`         | yes      | —       | Basic-auth username for `/url` endpoints      |
| `HTTP_PASSWORD`     | yes      | —       | Basic-auth password for `/url` endpoints      |

In Docker, `HTTP_ADDRESS` is set automatically from Railway's `PORT` (see `Dockerfile` `CMD`).