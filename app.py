from flask import Flask, request, Response, jsonify, send_file
import json
import requests
import io
import random
import datetime

import config

last = datetime.datetime.now()
app = Flask(__name__)



n_per_page = 10

@app.route('/', methods=['POST'])
def query():
    global last
    if datetime.datetime.now() - last < datetime.timedelta(seconds=5):
        return jsonify({"code": 429, "msg":"too many requests "}), 429
    last = datetime.datetime.now()

    # args
    keyword = request.form.get('pq')
    # start_index = request.form.get('start')
    secret = request.form.get('key')
    if keyword is None or secret != "123321":
        return jsonify({"code":401, "msg":"miss args"}), 401


    pics = Resources.query(keyword)
    n = len(pics)
    if n > 0:
        i = random.randint(0, n-1)
        pic = pics[i]
        # Resources.incr_use(pic["url"])
        return jsonify({"code":0, "msg": [pic["url"]]})

    # search request
    url = "https://www.googleapis.com/customsearch/v1"
    query_string = {
        "key": config.key,
        "cx": config.cx,
        "searchType": "image",
        "num": n_per_page,
        "q": keyword
        }
    # if start_index:
        # query_string["start"] = start_index
    resp = requests.request("GET", url, params=query_string)
    json_data = json.loads(resp.text)
    search_info =  'About %s results'%json_data['searchInformation']['formattedTotalResults'] 
    print(search_info)

    items = json_data.get('items', [])
    if len(items) == 0:
        print(json.dumps(json_data))
        return jsonify({"code": 404, "msg":"not found"}), 404

    start = json_data["queries"]["request"][0]["startIndex"]
    count = json_data["queries"]["request"][0]["count"]
    total = json_data["queries"]["request"][0]["totalResults"]
    search_id = SearchHistory.add(
        keyword,
        start,
        start+count,
        total,
        json.dumps(json_data))
    for item in items:
        Resources.add(item["link"], "", search_id)
    
    return jsonify({"code":0, "msg": [items[0]["link"]]})


@app.route('/pic', methods=['POST'])
def get():
    url = request.form.get('url')
    # start_index = request.form.get('start')
    secret = request.form.get('key')
    if url is None or secret != "456654":
        return jsonify({"code":401, "msg":"miss args"}), 401

    resp = requests.get(url)
    if resp.status_code != 200:
        return jsonify({"code": 405, "msg":"request failed"}), 405

    img = io.BytesIO(resp.content)
    content_type = resp.headers['content-type']
    return send_file(img, mimetype=content_type)
   
@app.errorhandler(404)
def not_found(error):
    return jsonify({"code": 404, "msg": "not found"}), 404

if __name__ == '__main__':
    app.run(host="0.0.0.0", port=config.port)