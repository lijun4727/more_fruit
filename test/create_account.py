#!python3
import requests
import ssl

ssl._create_default_https_context = ssl._create_unverified_context

url = 'https://127.0.0.1:8080/account/register'
params = {"user_name":"abcdedddd","password":"472780330","identity":"610425198710063911","phone":"13651884967"}
r = requests.post(url, json=params, verify=False).json()
print(r)