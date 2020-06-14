import datetime
import hashlib
import io
import json
import logging
import mimetypes
import os
import random

from PIL import Image
from flask import Flask
from flask import Response
from flask import jsonify
from flask import request
from flask import send_file
from ratelimit import limits
import requests

import config
from models import Resources
from models import SearchHistory

app = Flask(__name__)
n_per_page = 10
download_path = "download"


@app.route("/s", methods=["POST"])
@limits(calls=1, period=5)
def query():
    # args
    keyword = request.form.get("q")
    # start_index = request.form.get('start')
    secret = request.form.get("k")
    if keyword is None or secret != "123321":
        return jsonify({"code": 401, "msg": "miss args"}), 401

    pics = Resources.query_by_keyword(keyword)
    n = len(pics)
    if n > 0:
        # i = random.randint(0, n - 1)
        # pic = pics[i]
        # Resources.incr_use(pic["url"])
        return jsonify({"code": 0, "msg": [i["digest"] for i in pics]})

    # search request
    url = "https://www.googleapis.com/customsearch/v1"
    query_string = {"key": config.key, "cx": config.cx, "searchType": "image", "num": n_per_page, "q": keyword}
    # if start_index:
    # query_string["start"] = start_index
    resp = requests.request("GET", url, params=query_string)
    json_data = json.loads(resp.text)
    search_info = "About %s results" % json_data["searchInformation"]["formattedTotalResults"]
    print(search_info)

    items = json_data.get("items", [])
    if len(items) == 0:
        print(json.dumps(json_data))
        return jsonify({"code": 404, "msg": "not found"}), 404

    start = json_data["queries"]["request"][0]["startIndex"]
    count = json_data["queries"]["request"][0]["count"]
    total = json_data["queries"]["request"][0]["totalResults"]
    search_id = SearchHistory.add(keyword, start, start + count, total, json.dumps(json_data))
    pic_md5s = []
    for item in items:
        url = item["link"]
        filename = download_pic(url, download_path)
        if filename is None:
            continue
        s = filename.split(".")
        Resources.add(url, s[0], s[1], search_id)
        pic_md5s.append(s[0])

    return jsonify({"code": 0, "msg": pic_md5s})


@app.route("/pic", methods=["GET", "POST"])
@limits(calls=15, period=60)
def get():
    args = request.args if request.method == "GET" else request.form
    digest = args.get("id")
    size = args.get("size", 0)
    if size and not size.isdigit() or digest is None:
        return jsonify({"code": 401, "msg": "miss args"}), 401
    size = int(size)

    item = Resources.query_by_digest(digest)
    if item is None:
        return jsonify({"code": 404, "msg": "not found"}), 404

    filename = item["digest"] + "." + item["extname"]
    with Image.open(filename) as image:
        _format = image.format
        mime_type = image.get_format_mimetype()
        size_x, size_y = image.size

        if size and max(size_x, size_y) > size:
            ratio = max(size_x, size_y) / size
            new_size = (int(size_x / ratio), int(size_y / ratio))
            image = image.resize(new_size, resample=Image.LANCZOS, reducing_gap=30.0)

        data = io.BytesIO()
        image.save(data, format=_format, quality=92)  # quality only valide for jpeg
        data.seek(0)
    return send_file(data, mimetype=mime_type)


def download_pic(url, path):
    if not os.path.exists(path):
        os.mkdir(path)

    headers = {
        "user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36"
    }
    try:
        with requests.Session() as sess:
            resp = sess.get(url, headers=headers)
            if resp.status_code != 200:
                return

            mime = resp.headers["content-type"]
            extname = mimetypes.guess_extension(mime)
            h = hashlib.new("md5")
            h.update(resp.content)
            digest = h.hexdigest()
            filename = digest + extname
            Image.frombytes
            with open(path + "/" + filename, "wb") as f:
                for chunk in resp:
                    f.write(chunk)
    except Exception as ex:
        print(ex)
        return
    return filename


@app.errorhandler(404)
def not_found(error):
    return jsonify({"code": 404, "msg": "not found"}), 404


if __name__ == "__main__":
    logging.basicConfig(filename="error.log", level=logging.INFO)
    app.run(host="0.0.0.0", port=config.port, debug=False)
