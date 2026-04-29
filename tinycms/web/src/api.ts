export type Page = {
  id:number
  slug:string
  path:string
  title:string
  meta_description:string
  content_type:'page'|'post'
  markdown?:string
  published:boolean
  created_at:string
  updated_at:string
}

const base = '/cms.v1.CMSService/'
async function rpc<T>(name: string, body: Record<string, unknown> = {}): Promise<T> {
  const r = await fetch(base + name, { method: 'POST', credentials: 'include', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
  if (!r.ok) throw new Error(await r.text())
  return r.json()
}
export const api = {
  listPages: () => rpc<{pages:Page[]}>('ListPages'),
  getPage: (slug:string) => rpc<{page:Page}>('GetPage', { slug }),
  savePage: (page: Partial<Page>) => rpc<{page:Page}>('SavePage', page),
  deletePage: (slug:string) => rpc<{ok:boolean}>('DeletePage', { slug }),
  uploadFile: (name:string, data:string) => rpc<{asset:{id:number; name:string; url:string; size:number}}>('UploadFile', { name, data })
}
