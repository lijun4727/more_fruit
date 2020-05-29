#!python3

import requests
import ssl

ssl._create_default_https_context = ssl._create_unverified_context

url = 'https://127.0.0.1:8080/order/create'
params = {
    "id":"1334445533353351133",
    "goods_id":1, 
    "quantity":1,
    "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyTmFtZSI6ImxpanVuIiwiaXAiOiIxMjMiLCJleHAiOjE1ODc2NDM4ODJ9.wXzXuE9bwAP_FsDMkvyHJdjsuavmzzn1zmqA4nyINBk"
}
r = requests.post(url, json=params, verify=False).json()
print(r)