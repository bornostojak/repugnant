import { FormEvent, useEffect, useMemo, useState } from 'react'
import { createRoot } from 'react-dom/client'
import './styles.css'

type Project = { slug: string; name: string }
type Article = { id: string; title: string; body: string; revision: number; short_id: string; category: string; tags: string; source_path: string; source_range: string; created_at: string }
type CreatedProject = { slug: string; api_key: string; api_url: string }

const json = <T,>(url: string, init?: RequestInit) => fetch(url, init).then(async response => {
  if (!response.ok) throw new Error(await response.text() || response.statusText)
  return response.json() as Promise<T>
})
const tags = (article: Article) => { try { return JSON.parse(article.tags || '[]') as string[] } catch { return [] } }

function App() {
  const [projects, setProjects] = useState<Project[]>([])
  const [active, setActive] = useState('')
  const [articles, setArticles] = useState<Article[]>([])
  const [selected, setSelected] = useState<Article | null>(null)
  const [revisions, setRevisions] = useState<Article[]>([])
  const [query, setQuery] = useState('')
  const [view, setView] = useState<'docs' | 'source'>('docs')
  const [newName, setNewName] = useState('')
  const [created, setCreated] = useState<CreatedProject | null>(null)
  const [error, setError] = useState('')
  const refreshProjects = () => json<Project[]>('/api/projects').then(items => { setProjects(items); if (!active && items[0]) setActive(items[0].slug) }).catch(e => setError(String(e)))
  const refreshArticles = () => { if (active) void json<Article[]>(`/api/projects/${active}/articles?q=${encodeURIComponent(query)}`).then(setArticles).catch(e => setError(String(e))) }
  useEffect(() => { void refreshProjects() }, [])
  useEffect(() => { refreshArticles() }, [active, query])
  const grouped = useMemo(() => articles.reduce<Record<string, Article[]>>((all, article) => { const key = article.category || 'Uncategorised'; (all[key] ||= []).push(article); return all }, {}), [articles])
  const select = (article: Article) => { setSelected(article); setView('docs'); json<Article[]>(`/api/projects/${active}/articles/${article.id}/revisions`).then(setRevisions).catch(e => setError(String(e))); history.replaceState({}, '', `/p/${active}/article/${article.id}/${article.revision}`) }
  const createProject = (event: FormEvent) => { event.preventDefault(); setError(''); json<CreatedProject>('/api/projects', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({name:newName}) }).then(result => { setCreated(result); setNewName(''); refreshProjects(); setActive(result.slug) }).catch(e => setError(String(e))) }
  const saveOrganization = (event: FormEvent<HTMLFormElement>) => { event.preventDefault(); if (!selected) return; const form = new FormData(event.currentTarget); const category = String(form.get('category') || ''); const nextTags = String(form.get('tags') || '').split(',').map(value => value.trim()).filter(Boolean); json<Article>(`/api/projects/${active}/articles/${selected.id}`, {method:'PATCH', headers:{'Content-Type':'application/json'}, body:JSON.stringify({category, tags:nextTags})}).then(updated => { setSelected(updated); refreshArticles() }).catch(e => setError(String(e))) }
  return <main className="shell">
    <header className="topbar"><a className="brand" href="/">rPg</a><span>Documentation that stays close to code</span><nav><button className={view==='docs'?'selected':''} onClick={() => setView('docs')}>Docs</button><button className={view==='source'?'selected':''} onClick={() => setView('source')}>Source</button></nav></header>
    {error && <p className="notice" role="alert">{error}</p>}
    <section className="home-intro"><div><p className="eyebrow">Project wiki</p><h1>Understand the codebase without leaving its structure behind.</h1><p>Browse published documentation, source locations, tags, categories, and revision history.</p></div><form className="new-project" onSubmit={createProject}><label>New project<input required value={newName} onChange={e => setNewName(e.target.value)} placeholder="My service" /></label><button type="submit">Create project</button></form></section>
    {created && <section className="credential" aria-live="polite"><strong>{created.slug} is ready.</strong> Copy this API key now; it is only returned on creation. <code>{created.api_key}</code><p>Add <code>project.api_url: {created.api_url}</code> and the key to that project’s <code>rpg.conf.yaml</code>.</p></section>}
    <section className="workspace">
      <aside className="sidebar"><label>Project<select value={active} onChange={e=>{setActive(e.target.value);setSelected(null)}}><option value="">Choose a project</option>{projects.map(project => <option value={project.slug} key={project.slug}>{project.name}</option>)}</select></label><label>Search<input value={query} onChange={e=>setQuery(e.target.value)} placeholder="Titles, contents, tags" /></label><div className="tree"><h2>Documentation</h2>{Object.entries(grouped).map(([category, items]) => <section key={category}><h3>{category.split('/').join(' › ')}</h3>{items.map(article => <button className={selected?.id===article.id?'active':''} onClick={() => select(article)} key={article.id}>{article.title}</button>)}</section>)}</div></aside>
      <section className="content">{!active ? <p>Select or create a project to begin.</p> : !selected ? <p>Select an article from the documentation tree.</p> : view === 'source' ? <SourceView article={selected} /> : <ArticleView article={selected} revisions={revisions} onRevision={setSelected} />}</section>
      {selected && <aside className="details"><h2>Organize</h2><form onSubmit={saveOrganization}><label>Category path<input name="category" defaultValue={selected.category} placeholder="Settings/Caching" /></label><label>Search tags<input name="tags" defaultValue={tags(selected).join(', ')} placeholder="cache, performance" /></label><button type="submit">Save organization</button></form><h2>History</h2><ol>{revisions.map(revision => <li key={revision.revision}><button onClick={() => setSelected(revision)}>Revision {revision.revision}</button><small>{new Date(revision.created_at).toLocaleString()}</small></li>)}</ol><p className="muted">Stable short link: /{selected.short_id}</p></aside>}
    </section>
  </main>
}

function ArticleView({ article, revisions, onRevision }: { article: Article; revisions: Article[]; onRevision: (article: Article) => void }) {
  return <article className="article"><p className="eyebrow">{article.category || 'Uncategorised'} · revision {article.revision}</p><h1>{article.title}</h1><div className="tag-row">{tags(article).map(tag => <span key={tag}>{tag}</span>)}</div><pre>{article.body}</pre>{revisions.length > 1 && <p className="muted">Viewing a revision? <button onClick={() => onRevision(revisions[0])}>Return to latest</button></p>}</article>
}
function SourceView({ article }: { article: Article }) { return <article className="article"><p className="eyebrow">Documented source</p><h1>{article.source_path || 'Source location unavailable'}</h1><p>{article.source_range ? `Lines ${article.source_range}` : 'Source range unavailable'} · The CLI keeps exact quote content in each revision.</p><pre>{article.body.match(/## Documented code[\s\S]*/)?.[0] || 'No quoted source block was recorded.'}</pre></article> }

createRoot(document.getElementById('root')!).render(<App />)
