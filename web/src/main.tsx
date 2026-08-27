import { createRoot } from 'react-dom/client'
import { useEffect, useState } from 'react'
import './styles.css'

type Project={slug:string;name:string}; type Article={id:string;title:string;body:string;revision:number;short_id:string}
function App() {
  const [projects,setProjects]=useState<Project[]>([]); const [active,setActive]=useState(''); const [articles,setArticles]=useState<Article[]>([]); const [query,setQuery]=useState('');
  useEffect(()=>{fetch('/api/projects').then(r=>r.json()).then(setProjects)},[])
  useEffect(()=>{if(active)fetch(`/api/projects/${active}/articles?q=${encodeURIComponent(query)}`).then(r=>r.json()).then(setArticles)},[active,query])
  return (
    <main className="app-shell">
      <header><p className="eyebrow">rePugnant wiki</p><h1>Documentation that stays close to code.</h1><p className="lede">Choose a project to explore its published articles.</p></header>
      <section className="workspace"><aside><h2>Projects</h2>{projects.map(p=><button className={p.slug===active?'active':''} onClick={()=>setActive(p.slug)} key={p.slug}>{p.name}</button>)}</aside><section><input aria-label="Search documentation" value={query} onChange={e=>setQuery(e.target.value)} placeholder="Search title and contents" />{active?<><h2>{active}</h2>{articles.map(a=><article key={a.id}><a href={`/p/${active}/article/${a.id}/${a.revision}`}><h3>{a.title}</h3></a><p>{a.body.slice(0,160)}</p><small>revision {a.revision} · /{a.short_id}</small></article>)}</>:<p>Select a project to begin.</p>}</section></section>
    </main>
  )
}

createRoot(document.getElementById('root')!).render(<App />)
