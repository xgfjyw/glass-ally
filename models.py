from peewee import Model
from peewee import CharField, AutoField, DateTimeField, IntegerField, TextField
from peewee import DoesNotExist
import datetime
from playhouse.db_url import connect

import config


db = connect(config.db_dsn)

class SearchHistory(Model):
    id = AutoField(primary_key=True)
    keyword = CharField(max_length=128)
    start = IntegerField(default=0)
    end = IntegerField(default=0)
    total = IntegerField(default=0)
    result = TextField(default="")
    create_at = DateTimeField(default=datetime.datetime.now)

    class Meta:
        database = db
        table_name = "search_history"

    @classmethod
    def add(cls, keyword, start, end, total, result):
        item = cls.create(keyword=keyword, start=start, end=end, total=total, result=result)
        item.save()
        return item.id

    @classmethod
    def query(cls, keyword):
        q = cls.select().where(cls.keyword == keyword)
        return [{
            "id": i.id,
            "start": i.start,
            "end": i.end,
            "result": i.result,
            "create_at": i.create_at
        } for i in q]

class Resources(Model):
    id = AutoField(primary_key=True)
    url = CharField(max_length=512)
    digest = CharField(max_length=128, default="") #, unique=True)
    path = CharField(max_length=256, default="")
    search_id = IntegerField(default=0)
    used = IntegerField(default=0)
    create_at = DateTimeField(default=datetime.datetime.now)

    class Meta:
        database = db

    @classmethod
    def add(cls, url, path, search_id):
        # save file?
        item = cls.create(url=url, search_id=search_id, path=path)
        item.save()
        return item.id

    @classmethod
    def query(cls, keyword):
        searches = SearchHistory.query(keyword)
        search_ids = [i["id"] for i in searches]
        q = cls.select().where(cls.search_id.in_(search_ids))
        return [{
            "url": i.url,
            "path": i.path,
            "used": i.used,
            "create_at": i.create_at
        } for i in q]

    @classmethod
    def incr_use(cls, url):
        cls.update({cls.used: cls.used+1}).where(cls.url == url).execute()

db.create_tables([SearchHistory, Resources])