# APIs

## GET 

endpoint: /ping 

responce:
```
{"message": "pong"}
```

## POST

endpoint: /login

needs: 

```
{
    "user":     "someName",
    "password": "somePass"
}
```
responce:
```
{
    "message": "Hello someName"
}
```