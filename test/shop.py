#!python3
import requests
import ssl

ssl._create_default_https_context = ssl._create_unverified_context

url = 'https://127.0.0.1:8080/shop/addgoods'
params = {
    "Goodses" :[
        {
            "shop_id":9,
            "name":"苹果",
            "desc":"陕西水果",
            "quantity":10000
        },
        {
            "shop_id":8,
            "name":"李子",
            "desc":"陕西水果",
            "quantity":10000
        },
        {
            "shop_id":10,
            "name":"桃",
            "desc":"陕西水果",
            "quantity":10000
        },
        {
            "shop_id":11,
            "name":"杏",
            "desc":"陕西水果",
            "quantity":10000
        },
    ],
    "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyTmFtZSI6ImxpanVuIiwiaXAiOiIxMjMiLCJleHAiOjE1ODc2NDM4ODJ9.wXzXuE9bwAP_FsDMkvyHJdjsuavmzzn1zmqA4nyINBk"
}
r = requests.post(url, json=params, verify=False).json()
print(r)

url = 'https://127.0.0.1:8080/shop/build'
params = {
    "name":"水果多多",
    "desc":"李军",
    "account_id":2, 
    "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyTmFtZSI6ImxpanVuIiwiaXAiOiIxMjMiLCJleHAiOjE1ODc2NDM4ODJ9.wXzXuE9bwAP_FsDMkvyHJdjsuavmzzn1zmqA4nyINBk"
}
r = requests.post(url, json=params, verify=False).json()
print(r)