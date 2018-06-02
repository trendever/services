A way to bypass instagram protection.

Dump SigKey and http requests by dynamic code injecting into instagram app via frida.

How to:
1. Set up [frida](https://www.frida.re/docs/home/) and connect it to your device/vm/emulator
2. Run fri.py
3. ???
4. PROFIT!

Output example:

```
---------------------------
SigKey found: '20a8b13b818d739c88fb5c2602e720aa2572dcdcc46d2fd3d4f6d5473fafad61'
---------------------------

POST https://i.instagram.com/api/v1/ads/graphql/
X-IG-Connection-Type: "Ethernet"
X-IG-Capabilities: "3brTPw=="
X-IG-App-ID: "567067343352427"
User-Agent: "Instagram 47.0.0.16.96 Android (23/6.0.1; 160dpi; 1024x720; innotek GmbH/Android-x86; VirtualBox; x86_64; android_x86_64; en_US; 110937463)"
Accept-Language: "en-US"
Connection: "Keep-Alive"
Cookie: "is_starred_enabled=yes; sessionid=IGSCc48588fb35f0c53bab36fcb6fa0b5203acf881a748cbc6e8730a87231a8377dc%3Ahrn4lylDOE3xWEpHDc6H9ZiznzDQePjP%3A%7B%22_auth_user_id%22%3A3556800165%2C%22_auth_user_backend%22%3A%22accounts.backends.CaseInsensitiveModelBackend%22%2C%22_auth_user_hash%22%3A%22%22%2C%22_platform%22%3A1%2C%22_token_ver%22%3A2%2C%22_token%22%3A%223556800165%3A8VDzm4FDBkRb9ZYSy5rjJAsVtONxGWZH%3A34bd96b6e671a7103956ef91c78485500d4a563255abc91a6ceab11a27bf305d%22%2C%22last_refreshed%22%3A1527901477.4413928986%7D; shbts=1527924316.7239575; mid=WxBVXwABAAEuSdkgo6kjovgszNiL; ds_user=betrok; ds_user_id=3556800165; csrftoken=AlwFdBlmM0hfRBEOtVyGbWkNnXUjQQLA; igfl=betrok; shbid=9964; mcd=3; ig_direct_region_hint=ASH; rur=FTW; urlgen={\"
Content-Type: "application/x-www-form-urlencoded; charset=UTF-8"
Content-Length: "261"

query_id=1403787716369569&locale=en_US&vc_policy=ads_viewer_context_policy&signed_body=5f6d65f4bf7aec00cc2c3b7c923c3dae9268453ceff1d0aa8b100972e20e32fa.&ig_sig_key_version=4&strip_nulls=true&strip_defaults=true&query_params=%7B%220%22%3A%22793736047497610%22%7D
```

SigKey will be shown only once per session because its extraction slow down all the things.
