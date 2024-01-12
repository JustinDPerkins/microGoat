V1_API_KEY="YOU_API_KEY_HERE"

test = V1_API_KEY.strip('"')

api_key = test.split(' ', 1)[1]

print(api_key)