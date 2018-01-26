
The base framework for oodle is mostly done so now I can implement various commands.

Gonna implement them in this order:
- `oodle!` (done)
- `.seen` (wip)
- `.tell` (did not start)
- `.timefor` (needs forum api)
- `.profile` (needs forum api)
- `title` (easy to implement but low-pri)

## Prososals for forum/oodle integration

## Webhooks

### New post
- Method: POST
- Endpoint: https://oodle.pacn.in/event/thread
- Headers: X-SECRET: 415f1e53-1f74-4a42-ba78-4d89c2cba13f
- Payload:
```
{
    "user": "pacninja",
    "title": "API for oodle",
    "summary": "First 100 characters or so of post body",
    "permalink": "https://oods.net/showthread.php?id=125"
}
```

## Forum API

### Profile
- Method: GET
- Endpoint: https://oods.net/api/profile?user=pacninja
- Headers: X-API-KEY: 415f1e53-1f74-4a42-ba78-4d89c2cba13f
- Payload:
```
{
    "user": "pacninja",
    "about": "First 1 or 2 lines from about",
    "social": ["https://github.com/godwhoa", "https://twitter.com/godwhoa"],
    "permalink": "https://oods.net/profile.php?user=pacninja",
    "tz": "Asia/Kolkata",
}
```
For: `tz` for `.timefor <user>` and rest for `.profile <user>`

Discuss:

- How we wanna do auth(basic api keys or something fancy?)
- Payload format
- Endpoints (made them up on the spot; if they can be better do tell)
