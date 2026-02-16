import json,uuid,sys,requests
vec=[0.0]*384
obj={"points":[{"id":str(uuid.uuid4()),"vector":vec,"payload":{"text":"hi","deleted":False}}]}
print('len', len(vec))
print(json.dumps(obj)[:200])
resp=requests.put('http://localhost:6333/collections/kb_chunks/points?wait=true',json=obj)
print(resp.status_code, resp.text[:200])
