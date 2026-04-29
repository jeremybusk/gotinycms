import React, { useEffect, useMemo, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { App as AntApp, Button, Card, ConfigProvider, Form, Input, Layout, List, Popconfirm, Select, Space, Switch, Tabs, Typography, Upload, message, theme } from 'antd'
import type { UploadProps } from 'antd'
import { MDXEditor, headingsPlugin, listsPlugin, quotePlugin, thematicBreakPlugin, markdownShortcutPlugin, toolbarPlugin, UndoRedo, BoldItalicUnderlineToggles, ListsToggle, BlockTypeSelect, CreateLink, InsertImage, imagePlugin, linkDialogPlugin, linkPlugin, codeBlockPlugin, codeMirrorPlugin, InsertCodeBlock, CodeToggle, ConditionalContents, ChangeCodeMirrorLanguage, Separator, InsertTable, tablePlugin, InsertThematicBreak } from '@mdxeditor/editor'
import '@mdxeditor/editor/style.css'
import './style.css'
import { api, Page, SiteSettings } from './api'

const palettes = {
  slate: { colorPrimary: '#2563eb', colorBgLayout: '#f4f7fb', colorText: '#172033', colorBorder: '#d8dee9' },
  forest: { colorPrimary: '#3b7a57', colorBgLayout: '#f7f5ef', colorText: '#263238', colorBorder: '#e7e0d2' },
  ember: { colorPrimary: '#b45309', colorBgLayout: '#fff7ed', colorText: '#292524', colorBorder: '#fed7aa' },
  mono: { colorPrimary: '#111827', colorBgLayout: '#f9fafb', colorText: '#111827', colorBorder: '#e5e7eb' }
}

type Palette = keyof typeof palettes

function slugify(s:string) { return s.toLowerCase().trim().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '') }
function pathify(s:string) {
  if (/^https?:\/\//i.test(s) || /^mailto:/i.test(s) || /^tel:/i.test(s)) return s.trim()
  const path = s.split('/').map(slugify).filter(Boolean).join('/')
  return path ? `/${path}` : '/'
}
function freshPage() {
  return { slug:'', path:'', title:'', meta_description:'', content_type:'page', markdown:'# Untitled\n', published:false } as Page
}
function newID() {
  return globalThis.crypto?.randomUUID?.() || `item-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function Root() {
  const [pages, setPages] = useState<Page[]>([])
  const [active, setActive] = useState<Page | null>(null)
  const [saving, setSaving] = useState(false)
  const [savingSettings, setSavingSettings] = useState(false)
  const [sourceMode, setSourceMode] = useState(false)
  const [editorRev, setEditorRev] = useState(0)
  const [palette, setPalette] = useState<Palette>('slate')
  const [adminDark, setAdminDark] = useState(false)
  const [form] = Form.useForm()
  const [settingsForm] = Form.useForm<SiteSettings>()
  const md = Form.useWatch('markdown', form) ?? ''
  const menuItems = Form.useWatch('menu', settingsForm) || []
  const cfg = useMemo(() => ({ token: palettes[palette], algorithm: adminDark ? theme.darkAlgorithm : theme.defaultAlgorithm }), [palette, adminDark])

  async function loadPages() {
    const r = await api.listPages()
    setPages(r.pages)
    if (!active && r.pages[0]) openPage(r.pages[0].slug)
  }
  async function loadSettings() {
    const r = await api.getSettings()
    settingsForm.setFieldsValue(r.settings)
  }
  async function openPage(slug:string) {
    const r = await api.getPage(slug)
    setActive(r.page)
    form.setFieldsValue(r.page)
    setEditorRev(rev => rev + 1)
  }
  async function savePage() {
    setSaving(true)
    try {
      const v = form.getFieldsValue()
      const r = await api.savePage({ ...v, slug: slugify(v.slug || v.title || ''), path: pathify(v.path || v.slug || v.title || '') })
      setActive(r.page)
      form.setFieldsValue(r.page)
      setEditorRev(rev => rev + 1)
      await loadPages()
      message.success('Page saved')
    } catch(e:any) {
      message.error(e.message)
    } finally {
      setSaving(false)
    }
  }
  async function saveSettings() {
    setSavingSettings(true)
    try {
      const values = settingsForm.getFieldsValue()
      const menu = (values.menu || []).map(item => ({ ...item, id: item.id || newID(), parent_id: item.parent_id || '', url: pathify(item.url || ''), enabled: item.enabled !== false })).filter(item => item.label && item.url)
      const r = await api.saveSettings({ ...values, menu })
      settingsForm.setFieldsValue(r.settings)
      message.success('Site settings saved')
    } catch(e:any) {
      message.error(e.message)
    } finally {
      setSavingSettings(false)
    }
  }
  async function removePage(slug:string) {
    await api.deletePage(slug)
    setActive(null)
    form.resetFields()
    await loadPages()
  }
  function newPage() {
    const p = freshPage()
    setActive(p)
    form.setFieldsValue(p)
    setSourceMode(false)
    setEditorRev(rev => rev + 1)
  }

  useEffect(() => {
    loadPages().catch(e => message.error(e.message))
    loadSettings().catch(e => message.error(e.message))
  }, [])

  async function upload(file: File) {
    const data = await new Promise<string>((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => resolve(String(reader.result))
      reader.onerror = () => reject(reader.error)
      reader.readAsDataURL(file)
    })
    return api.uploadFile(file.name, data)
  }
  async function uploadImageForEditor(file: File) {
    const r = await upload(file)
    return r.asset.url
  }

  const contentUploadProps: UploadProps = {
    showUploadList: false,
    beforeUpload(file) {
      upload(file).then(r => {
        form.setFieldValue('markdown', `${form.getFieldValue('markdown') || ''}\n\n![${file.name}](${r.asset.url})\n`)
        setEditorRev(rev => rev + 1)
        message.success('Uploaded')
      }).catch((e:any) => message.error(e.message))
      return false
    }
  }
  function settingsUploadProps(field: 'logo_url' | 'favicon_url'): UploadProps {
    return {
      showUploadList: false,
      accept: field === 'favicon_url' ? '.ico,.png,.jpg,.jpeg,.webp' : '.png,.jpg,.jpeg,.webp,.avif',
      beforeUpload(file) {
        upload(file).then(r => {
          settingsForm.setFieldValue(field, r.asset.url)
          message.success(field === 'logo_url' ? 'Logo uploaded' : 'Favicon uploaded')
        }).catch((e:any) => message.error(e.message))
        return false
      }
    }
  }

  return <ConfigProvider theme={cfg}><AntApp><Layout className="layout">
    <Layout.Sider className="sider" width={310} breakpoint="lg" collapsedWidth={0}>
      <div className="brand">TinyCMS</div>
      <Space wrap className="palettes">
        {Object.keys(palettes).map(p => <Button key={p} size="small" type={p===palette?'primary':'default'} onClick={() => setPalette(p as Palette)}>{p}</Button>)}
        <Switch checkedChildren="Dark" unCheckedChildren="Light" checked={adminDark} onChange={setAdminDark} />
      </Space>
      <Button block type="primary" onClick={newPage}>New page</Button>
      <List className="pages" dataSource={pages} renderItem={p => <List.Item className={active?.slug===p.slug?'selected':''} onClick={() => openPage(p.slug)}>
        <List.Item.Meta title={p.title} description={`${p.path || `/${p.slug}`}${p.content_type === 'post' ? ' · post' : ''}${p.published ? '' : ' · draft'}`} />
      </List.Item>} />
    </Layout.Sider>
    <Layout.Content className="content">
      <Tabs defaultActiveKey="content" items={[
        { key:'content', label:'Content', children:<Card className="editorCard">
          <Form form={form} layout="vertical" onFinish={savePage} initialValues={freshPage()}>
            <Space className="topbar" align="start">
              <Typography.Title level={3}>{active?.id ? 'Edit page' : 'New page'}</Typography.Title>
              <Space wrap>
                <Switch checkedChildren="Markdown" unCheckedChildren="Editor" checked={sourceMode} onChange={setSourceMode} />
                {active?.slug && active.slug !== 'home' && <Popconfirm title="Delete page?" onConfirm={() => removePage(active.slug)}><Button danger>Delete</Button></Popconfirm>}
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
            <Upload {...contentUploadProps}><Button>Upload image/file and insert Markdown link</Button></Upload>
            <Form.Item name="markdown" label="Body" className="mdField">
              {sourceMode
                ? <Input.TextArea rows={22} className="sourceEditor" value={md} onChange={e => form.setFieldValue('markdown', e.target.value)} />
                : <MDXEditor key={`${active?.slug || 'new'}-${active?.updated_at || ''}-${editorRev}`} markdown={md} onChange={v => form.setFieldValue('markdown', v)} plugins={[
                  headingsPlugin(),
                  listsPlugin(),
                  quotePlugin(),
                  thematicBreakPlugin(),
                  linkPlugin(),
                  linkDialogPlugin(),
                  imagePlugin({ imageUploadHandler: uploadImageForEditor, imageAutocompleteSuggestions: ['/uploads/'] }),
                  tablePlugin(),
                  codeBlockPlugin({ defaultCodeBlockLanguage: 'text' }),
                  codeMirrorPlugin({ codeBlockLanguages: { text: 'Plain text', markdown: 'Markdown', javascript: 'JavaScript', typescript: 'TypeScript', jsx: 'JSX', tsx: 'TSX', html: 'HTML', css: 'CSS', json: 'JSON', bash: 'Shell', go: 'Go', sql: 'SQL', yaml: 'YAML', mermaid: 'Mermaid diagram' } }),
                  markdownShortcutPlugin(),
                  toolbarPlugin({toolbarContents: () => <ConditionalContents options={[
                    { when: editor => editor?.editorType === 'codeblock', contents: () => <ChangeCodeMirrorLanguage /> },
                    { fallback: () => <><UndoRedo /><Separator /><BoldItalicUnderlineToggles /><CodeToggle /><ListsToggle /><BlockTypeSelect /><Separator /><CreateLink /><InsertImage /><Separator /><InsertTable /><InsertThematicBreak /><InsertCodeBlock /></> }
                  ]} />})
                ]} />}
            </Form.Item>
          </Form>
        </Card> },
        { key:'site', label:'Site', children:<Card className="editorCard">
          <Form form={settingsForm} layout="vertical" onFinish={saveSettings} initialValues={{site_name:'TinyCMS', default_theme:'light', footer_markdown:'', logo_enabled:true, favicon_enabled:true, menu_enabled:true, footer_enabled:true, theme_toggle_enabled:true, icons_enabled:true, menu:[{id:'home', parent_id:'', label:'Home', url:'/', external:false, enabled:true}]}}>
            <Space className="topbar" align="start">
              <div>
                <Typography.Title level={3}>Site settings</Typography.Title>
                <Typography.Text type="secondary">Logo, favicon, top menu, footer, and public light/dark default.</Typography.Text>
              </div>
              <Button type="primary" htmlType="submit" loading={savingSettings}>Save site</Button>
            </Space>
            <Space className="switchGrid" wrap>
              <Form.Item name="logo_enabled" label="Logo" valuePropName="checked"><Switch /></Form.Item>
              <Form.Item name="favicon_enabled" label="Favicon" valuePropName="checked"><Switch /></Form.Item>
              <Form.Item name="menu_enabled" label="Top menu" valuePropName="checked"><Switch /></Form.Item>
              <Form.Item name="footer_enabled" label="Footer" valuePropName="checked"><Switch /></Form.Item>
              <Form.Item name="theme_toggle_enabled" label="Guest theme toggle" valuePropName="checked"><Switch /></Form.Item>
              <Form.Item name="icons_enabled" label="Font Awesome icons" valuePropName="checked"><Switch /></Form.Item>
            </Space>
            <Form.Item name="site_name" label="Site name" rules={[{required:true}]}><Input /></Form.Item>
            <Space className="assetGrid" align="start">
              <Form.Item name="logo_url" label="Logo URL"><Input placeholder="/uploads/..." /></Form.Item>
              <Upload {...settingsUploadProps('logo_url')}><Button>Upload logo</Button></Upload>
              <Form.Item name="favicon_url" label="Favicon URL"><Input placeholder="/uploads/..." /></Form.Item>
              <Upload {...settingsUploadProps('favicon_url')}><Button>Upload favicon</Button></Upload>
            </Space>
            <Form.Item name="default_theme" label="Public default theme"><Select options={[{label:'Light', value:'light'}, {label:'Dark', value:'dark'}]} /></Form.Item>
            <Typography.Title level={4}>Top menu</Typography.Title>
            <Form.List name="menu">{(fields, { add, remove }) => <>
              {fields.map(field => <Space key={field.key} className="menuRow" align="start">
                <Form.Item {...field} name={[field.name, 'id']} hidden><Input /></Form.Item>
                <Form.Item {...field} name={[field.name, 'label']} label="Label" rules={[{required:true}]}><Input placeholder="About" /></Form.Item>
                <Form.Item {...field} name={[field.name, 'url']} label="URL" rules={[{required:true}]}><Input placeholder="/about or https://..." /></Form.Item>
                <Form.Item {...field} name={[field.name, 'parent_id']} label="Parent"><Select allowClear placeholder="Top level" options={(menuItems || []).filter((item, i) => i !== field.name && item?.id).map(item => ({label: item.label || item.url || item.id, value: item.id}))} /></Form.Item>
                <Form.Item {...field} name={[field.name, 'external']} label="External" valuePropName="checked"><Switch /></Form.Item>
                <Form.Item {...field} name={[field.name, 'enabled']} label="Enabled" valuePropName="checked"><Switch /></Form.Item>
                <Button danger onClick={() => remove(field.name)}>Remove</Button>
              </Space>)}
              <Button onClick={() => add({id:newID(), parent_id:'', label:'', url:'/', external:false, enabled:true})}>Add menu item</Button>
            </>}</Form.List>
            <Form.Item name="footer_markdown" label="Global footer Markdown" className="footerField"><Input.TextArea rows={6} placeholder="© 2026 Your Company. All rights reserved." /></Form.Item>
          </Form>
        </Card> }
      ]} />
    </Layout.Content>
  </Layout></AntApp></ConfigProvider>
}

createRoot(document.getElementById('root')!).render(<Root />)
