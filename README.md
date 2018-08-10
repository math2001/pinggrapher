# Ping graph

Graph the pings

## Example

Ping the local router, and only prints the milliseconds

```sh
$ ping 192.168.0.1 | awk  '/from/ { split($7, resArr, "="); print resArr[2] }'
```


