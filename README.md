# web

API for making HTTP REST requests.

## Usage (plain text)

If the response is plain text, use `Request`.

```golang
import (
  "github.com/erikbryant/web"
)

json, err := web.RequestJSON("https://www.vesselfinder.com/api/pub/vesselsonmap", {})
if err != nil {
  return err
}
```

## Usage (JSON)

If the response is JSON, use `RequestJSON`.

```golang
import (
  "fmt"
  "github.com/erikbryant/web"
)

response, err := web.RequestJSON("https://api.ipstack.com/check", {})
if err != nil {
  return err
}

fmt.Println(response)
// response = {
//   "success": false,
//   "error": {
//     "code": 101,
//     "type": "missing_access_key",
//     "info": "You have not supplied an API Access Key. [Required format: access_key=YOUR_ACCESS_KEY]"
//   }
// }

error := response["error"].([]interface{})[0].(map[string]interface{})
code := web.ToInt(error["code"])

fmt.Println(code)
// 101
```

## Setting Headers

You can pass in a map of headers that you wish to set.

```golang
import (
  "fmt"
  "github.com/erikbryant/web"
)

url := "https://www.marinetraffic.com/getData/get_data_json_4/z:14/X:1309/Y:3165/station:0"
headers := map[string]string{
  "user-agent":       "ship-ahoy",
  "x-requested-with": "XMLHttpRequest",
  "vessel-image":     "001609ab6d06a620f459d4a1fd65f1315f11",
}

response, err := web.RequestJSON(url, headers)
if err != nil {
  return err
}

fmt.Println(response)
// response = {
//   "type":1,
//   "data":{
//     "rows":[
//       {"LAT":"37.80998","LON":"-122.4215","SPEED":"0", ...},
//       ...
//     ]
//   }
// }
```
