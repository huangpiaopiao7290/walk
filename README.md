# walk
A project of walk and look, based on go and react

## content
There are two sections about Walk, that are tag and grasp. The tag part is suppotred by Amap, and you can  mark locations where you have gone in your earth model. In another grasp part, recording what you're looking here in each of your memeory tabs.


## Program Introduction
### env  
- wsl2 arch
- go 1.24.0
- etcd
- protocol buffers
- mariadb
- redis



### Docker
- base
- map-service
- user-service

### map-service
configration introduction `path: walk/apps/map/config/map-service.yaml`:
```yaml
amap:
  web:
    key: your_amap_web_key
    signature: your_amap_web_signature
    geocodeBaseURL: https://restapi.amap.com/v3/geocode/geo
    staticMapBaseURL: https://restapi.amap.com/v3/staticmap
    ...
  web_js:
    key: your_amap_web_js_key
    private_key: your_amap_web_js_private_key

 ...
```


### user-service


