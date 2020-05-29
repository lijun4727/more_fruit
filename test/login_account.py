#!python3
import requests
import ssl

ssl._create_default_https_context = ssl._create_unverified_context

url = 'https://127.0.0.1:8080/account/login'
params = {"user_name":"lijun","password":"12"}
r = requests.post(url, json=params, verify=False).json()
print(r)