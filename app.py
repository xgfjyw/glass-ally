import datetime
import io
import json
import random
import os

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


@app.route("/s", methods=["POST"])
@limits(calls=1, period=5)
def query():
    # args
    keyword = request.form.get("q")
    # start_index = request.form.get('start')
    secret = request.form.get("k")
    if keyword is None or secret != "123321":
        return jsonify({"code": 401, "msg": "miss args"}), 401

    pics = Resources.query(keyword)
    n = len(pics)
    if n > 0:
        i = random.randint(0, n - 1)
        pic = pics[i]
        # Resources.incr_use(pic["url"])
        return jsonify({"code": 0, "msg": [pic["url"]]})

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
    for item in items:
        Resources.add(item["link"], "", search_id)

    return jsonify({"code": 0, "msg": [items[0]["link"]]})


@app.route("/pic", methods=["POST"])
@limits(calls=15, period=60)
def get():
    url = request.form.get("url")
    # start_index = request.form.get('start')
    secret = request.form.get("key")
    if url is None or secret != "456654":
        return jsonify({"code": 401, "msg": "miss args"}), 401


return jsonify({"code": 405, "msg": "request failed"}), 405

    img = io.BytesIO(resp.content)
    content_type = resp.headers["content-type"]
    return send_file(img, mimetype=content_type)

def get_pic_file(url, path, filename):
    if os.path.exists(path + "/" + filename):
        with open(path + "/" + filename) as f:
            data = f.read()
        return data
    if not os.exit(path):
        os.mkdir(path)

    headers = {"user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36"}
    with requests.Session() as sess:
        sess.get(url, header=headers)
        if resp.status_code != 200:
            return
        img = io.BytesIO(resp.content)
    with open(path + "/" + filename, "")
        

@app.errorhandler(404)
def not_found(error):
    return jsonify({"code": 404, "msg": "not found"}), 404


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=config.port)
