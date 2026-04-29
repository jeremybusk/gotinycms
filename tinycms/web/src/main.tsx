import React, { useEffect, useMemo, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { App as AntApp, Button, Card, ConfigProvider, Form, Input, Layout, List, Popconfirm, Select, Space, Switch, Typography, Upload, message, theme } from 'antd'
import type { UploadProps } from 'antd'
import { MDXEditor, headingsPlugin, listsPlugin, quotePlugin, thematicBreakPlugin, markdownShortcutPlugin, toolbarPlugin, UndoRedo, BoldItalicUnderlineToggles, ListsToggle, BlockTypeSelect, CreateLink, InsertImage } from '@mdxeditor/editor'
import '@mdxeditor/editor/style.css'
import './style.css'
import { api, Page } from './api'

const palettes = {
  forest: { colorPrimary: '#3b7a57', colorBgLayout: '#f7f5ef', colorText: '#263238', colorBorder: '#e7e0d2' },
  slate: { colorPrimary: '#2563eb', colorBgLayout: '#f6f8fb', colorText: '#1f2937', colorBorder: '#d8dee9' },
  ember: { colorPrimary: '#b45309', colorBgLayout: '#fff7ed', colorText: '#292524', colorBorder: '#fed7aa' },
  plum: { colorPrimary: '#7c3aed', colorBgLayout: '#faf5ff', colorText: '#2e1065', colorBorder: '#e9d5ff' },
  mono: { colorPrimary: '#111827', colorBgLayout: '#f9fafb', colorText: '#111827', colorBorder: '#e5e7eb' }
}

type Palette = keyof typeof palettes

function slugify(s:string) { return s.toLowerCase().trim().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '') }
function pathify(s:string) {
  const path = s.split('/').map(slugify).filter(Boolean).join('/')
  return path ? `/${path}` : '/'
}

function Root() {
  const [pages, setPages] = useState<Page[]>([])
  const [active, setActive] = useState<Page | null>(null)
  const [saving, setSaving] = useState(false)
  const [palette, setPalette] = useState<Palette>('forest')
  const [form] = Form.useForm()
  const md = Form.useWatch('markdown', form) ?? ''
  const cfg = useMemo(() => ({ token: palettes[palette], algorithm: theme.defaultAlgorithm }), [palette])

  async function load() { const r = await api.listPages(); setPages(r.pages); if (!active && r.pages[0]) open(r.pages[0].slug) }
  async function open(slug:string) { const r = await api.getPage(slug); setActive(r.page); form.setFieldsValue(r.page) }
  async function save() { setSaving(true); try { const v = form.getFieldsValue(); const r = await api.savePage({ ...v, slug: slugify(v.slug || v.title || ''), path: pathify(v.path || v.slug || v.title || '') }); setActive(r.page); form.setFieldsValue(r.page); await load(); message.success('Saved') } catch(e:any) { message.error(e.message) } finally { setSaving(false) } }
  async function remove(slug:string) { await api.deletePage(slug); setActive(null); form.resetFields(); await load() }
  function fresh() { const p = { slug:'', path:'', title:'', meta_description:'', content_type:'page', markdown:'# Untitled\n', published:false } as Page; setActive(p); form.setFieldsValue(p) }

  useEffect(() => { load().catch(e => message.error(e.message)) }, [])

  const uploadProps: UploadProps = {
    showUploadList: false,
    beforeUpload(file) {
      const reader = new FileReader()
      reader.onload = async () => {
        try { const r = await api.uploadFile(file.name, String(reader.result)); form.setFieldValue('markdown', `${form.getFieldValue('markdown') || ''}\n\n![${file.name}](${r.asset.url})\n`); message.success('Uploaded') } catch(e:any) { message.error(e.message) }
      }
      reader.readAsDataURL(file)
      return false
    }
  }

  return <ConfigProvider theme={cfg}><AntApp><Layout className="layout">
    <Layout.Sider className="sider" width={300} breakpoint="lg" collapsedWidth={0}>
      <div className="brand">TinyCMS</div>
      <Space wrap className="palettes">{Object.keys(palettes).map(p => <Button key={p} size="small" type={p===palette?'primary':'default'} onClick={() => setPalette(p as Palette)}>{p}</Button>)}</Space>
      <Button block type="primary" onClick={fresh}>New page</Button>
        <List className="pages" dataSource={pages} renderItem={p => <List.Item className={active?.slug===p.slug?'selected':''} onClick={() => open(p.slug)}>
        <List.Item.Meta title={p.title} description={`${p.path || `/${p.slug}`}${p.content_type === 'post' ? ' · post' : ''}${p.published ? '' : ' · draft'}`} />
      </List.Item>} />
    </Layout.Sider>
    <Layout.Content className="content">
      <Card className="editorCard">
        <Form form={form} layout="vertical" onFinish={save} initialValues={{published:false}}>
          <Space className="topbar" align="start">
            <Typography.Title level={3}>{active?.id ? 'Edit page' : 'New page'}</Typography.Title>
            <Space>
              {active?.slug && active.slug !== 'home' && <Popconfirm title="Delete page?" onConfirm={() => remove(active.slug)}><Button danger>Delete</Button></Popconfirm>}
              <Button href={active?.path || '/'} target="_blank">View</Button>
              <Button type="primary" htmlType="submit" loading={saving}>Save</Button>
            </Space>
          </Space>
          <Form.Item name="title" label="Title" rules={[{required:true}]}><Input onBlur={() => {
            const title = form.getFieldValue('title') || ''
            if (!form.getFieldValue('slug')) form.setFieldValue('slug', slugify(title))
            if (!form.getFieldValue('path')) form.setFieldValue('path', pathify(title))
          }} /></Form.Item>
          <Space className="routeGrid" align="start">
            <Form.Item name="slug" label="Admin slug" rules={[{required:true}]}><Input addonBefore="id:" /></Form.Item>
            <Form.Item name="path" label="Public route / SEO URL" rules={[{required:true}]}><Input placeholder="/about/company" /></Form.Item>
            <Form.Item name="content_type" label="Type" rules={[{required:true}]}><Select options={[{label:'Page', value:'page'}, {label:'Post', value:'post'}]} /></Form.Item>
            <Form.Item name="published" label="Published" valuePropName="checked"><Switch /></Form.Item>
          </Space>
          <Form.Item name="meta_description" label="SEO description"><Input.TextArea rows={2} maxLength={180} showCount placeholder="Short search/social description for this route." /></Form.Item>
          <Upload {...uploadProps}><Button>Upload image/file and insert Markdown link</Button></Upload>
          <Form.Item name="markdown" label="Markdown" className="mdField"><MDXEditor markdown={md} onChange={v => form.setFieldValue('markdown', v)} plugins={[headingsPlugin(), listsPlugin(), quotePlugin(), thematicBreakPlugin(), markdownShortcutPlugin(), toolbarPlugin({toolbarContents: () => <><UndoRedo /><BoldItalicUnderlineToggles /><ListsToggle /><BlockTypeSelect /><CreateLink /><InsertImage /></>})]} /></Form.Item>
        </Form>
      </Card>
    </Layout.Content>
  </Layout></AntApp></ConfigProvider>
}

createRoot(document.getElementById('root')!).render(<Root />)
