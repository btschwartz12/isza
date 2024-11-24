import sys
from instagrapi import Client

username = sys.argv[1]
password = sys.argv[2]
paths = sys.argv[3].split(',')
caption_file = sys.argv[4]
test = sys.argv[5] == 'true'

with open(caption_file, 'r') as f:
    caption = f.read()

if test:
    print("Test mode: not posting to Instagram")
    print(f"Username: {username}")
    print(f"Password: {password}")
    print(f"Paths: {paths}")
    print(f"Caption: {caption}")
    exit(0)
cl = Client()

if len(paths) == 0:
    raise Exception("No paths provided")

attempts = 0

try:
    cl.login(username, password)
    if len(paths) == 1:
        media = cl.photo_upload(path=paths[0], caption=caption)
    else:
        media = cl.album_upload(paths, caption=caption)
    cl.logout()
    print(media)
except Exception as e:
    print(f"Error: {e}")
    exit(1)
