// Reference only (Figma Make export). Not imported by the app; see frontend/AGENTS.md.
import { useEffect, useState, type ReactNode } from 'react'

type IconName =
  | 'box'
  | 'check'
  | 'chevron'
  | 'clock'
  | 'copy'
  | 'download'
  | 'external'
  | 'file'
  | 'info'
  | 'layers'
  | 'link'
  | 'map'
  | 'moon'
  | 'route'
  | 'scan'
  | 'shield'
  | 'sun'

const paths: Record<IconName, ReactNode> = {
  box: <><path d="m4.5 7.5 7.5 4 7.5-4M12 21V11.5"/><path d="m4.5 7.5 7.5-4 7.5 4v9l-7.5 4.5-7.5-4.5Z"/></>,
  check: <path d="m5 12.5 4.3 4.3L19.5 6.5"/>,
  chevron: <path d="m9 6 6 6-6 6"/>,
  clock: <><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/></>,
  copy: <><rect x="8" y="8" width="11" height="11" rx="2"/><path d="M16 8V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h2"/></>,
  download: <><path d="M12 3v12m0 0 4-4m-4 4-4-4"/><path d="M5 20h14"/></>,
  external: <><path d="M14 4h6v6M20 4l-9 9"/><path d="M18 13v5a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h5"/></>,
  file: <><path d="M6 3h8l4 4v14H6z"/><path d="M14 3v5h5M9 13h6M9 17h4"/></>,
  info: <><circle cx="12" cy="12" r="9"/><path d="M12 11v6M12 7.5v.2"/></>,
  layers: <><path d="m12 3 9 5-9 5-9-5z"/><path d="m3 12 9 5 9-5M3 16l9 5 9-5"/></>,
  link: <><path d="M9.5 14.5 14.5 9"/><path d="M7.5 17H6a4 4 0 0 1 0-8h3M16.5 7H18a4 4 0 0 1 0 8h-3"/></>,
  map: <><path d="m3 6 6-3 6 3 6-3v15l-6 3-6-3-6 3z"/><path d="M9 3v15M15 6v15"/></>,
  moon: <path d="M20 15.2A8.5 8.5 0 0 1 8.8 4a8.6 8.6 0 1 0 11.2 11.2Z"/>,
  route: <><circle cx="6" cy="18" r="2"/><circle cx="18" cy="6" r="2"/><path d="M8 18h3a3 3 0 0 0 0-6H9a3 3 0 0 1 0-6h7"/></>,
  scan: <><path d="M4 9V5a1 1 0 0 1 1-1h4M15 4h4a1 1 0 0 1 1 1v4M20 15v4a1 1 0 0 1-1 1h-4M9 20H5a1 1 0 0 1-1-1v-4"/><path d="M7 12h10"/></>,
  shield: <><path d="M12 3 20 6v5c0 5-3.4 8.3-8 10-4.6-1.7-8-5-8-10V6z"/><path d="m8.5 12 2.2 2.2 4.8-5"/></>,
  sun: <><circle cx="12" cy="12" r="3.5"/><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"/></>,
}

function Icon({ name, size = 18 }: { name: IconName; size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      {paths[name]}
    </svg>
  )
}

function Logo() {
  return (
    <div className="brand">
      <span className="brand-mark"><Icon name="link" size={21} /></span>
      <span>ChainTrack</span>
    </div>
  )
}

const checks = [
  ['Origin verified', 'Licensed facility and owner confirmed'],
  ['Chain records match', 'All anchored records are unchanged'],
  ['Mass balance passed', 'Quantities reconcile across transforms'],
  ['Certificates valid', 'Issuers and validity windows confirmed'],
]

const events = [
  { date: '08 Jun 2026', title: 'Coffee harvested', place: 'Nyamagabe Highlands Farm', kind: 'Origin', icon: 'box' as IconName },
  { date: '10 Jun 2026', title: 'Washed & processed', place: 'Kivu Belt Washing Station', kind: 'Transform', icon: 'layers' as IconName },
  { date: '14 Jun 2026', title: 'Custody accepted', place: 'Horizon Coffee Exporters', kind: 'Handover', icon: 'check' as IconName },
  { date: '19 Jun 2026', title: 'Export cleared', place: 'Mombasa Port, Kenya', kind: 'Export', icon: 'route' as IconName },
]

function PublicVerification({ onBusiness }: { onBusiness: () => void }) {
  const [dark, setDark] = useState(false)
  const [activeTab, setActiveTab] = useState<'journey' | 'genealogy'>('journey')
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    document.documentElement.dataset.theme = dark ? 'dark' : 'light'
  }, [dark])

  const copyId = () => {
    navigator.clipboard?.writeText('CT-RW-2026-04821')
    setCopied(true)
    window.setTimeout(() => setCopied(false), 1500)
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="topbar-inner">
          <Logo />
          <div className="header-actions">
            <button className="icon-button" onClick={() => setDark(!dark)} aria-label="Toggle color mode">
              <Icon name={dark ? 'sun' : 'moon'} />
            </button>
            <button className="outline-button"><Icon name="scan" /> Scan another</button>
            <button className="text-button" onClick={onBusiness}>Business sign in <Icon name="chevron" size={15} /></button>
          </div>
        </div>
      </header>

      <main>
        <section className="passport-hero">
          <div className="hero-glow" />
          <div className="page-container hero-inner">
            <div className="breadcrumb"><span>Public verification</span><Icon name="chevron" size={13}/><strong>CT-RW-2026-04821</strong></div>
            <div className="hero-grid">
              <div className="verdict">
                <div className="verdict-icon"><Icon name="shield" size={31}/></div>
                <div>
                  <div className="eyebrow">Verification result</div>
                  <h1>Verified</h1>
                  <p>This product’s origin and journey have been independently verified.</p>
                </div>
              </div>
              <div className="verification-meta">
                <div>
                  <span>Verified on</span>
                  <strong>24 June 2026, 14:32 CAT</strong>
                </div>
                <div>
                  <span>Network record</span>
                  <strong className="teal-text"><i className="status-dot"/> All records anchored</strong>
                </div>
              </div>
            </div>
          </div>
        </section>

        <div className="page-container content-grid">
          <div className="primary-column">
            <section className="card product-card">
              <div className="product-image">
                <div className="bean bean-one"/>
                <div className="bean bean-two"/>
                <div className="bean bean-three"/>
                <div className="origin-tag">RWANDA · SPECIALTY</div>
              </div>
              <div className="product-details">
                <div className="label-row">
                  <span className="pill neutral">Arabica coffee</span>
                  <span className="pill verified"><Icon name="check" size={13}/> Active passport</span>
                </div>
                <h2>Nyamagabe Bourbon</h2>
                <p className="product-description">Fully washed red bourbon · Grade A</p>
                <div className="detail-grid">
                  <div><span>Lot ID</span><button className="inline-copy" onClick={copyId}><code>CT-RW-2026-04821</code><Icon name={copied ? 'check' : 'copy'} size={14}/></button></div>
                  <div><span>Net quantity</span><strong>2,400 kg</strong></div>
                  <div><span>Harvest period</span><strong>May – June 2026</strong></div>
                  <div><span>Current holder</span><strong>Aster Roasters Ltd.</strong></div>
                </div>
              </div>
            </section>

            <section className="card journey-card">
              <div className="tabs" role="tablist">
                <button className={activeTab === 'journey' ? 'tab active' : 'tab'} onClick={() => setActiveTab('journey')}><Icon name="route"/> Journey</button>
                <button className={activeTab === 'genealogy' ? 'tab active' : 'tab'} onClick={() => setActiveTab('genealogy')}><Icon name="layers"/> Genealogy</button>
              </div>
              {activeTab === 'journey' ? (
                <div className="journey-layout">
                  <div className="timeline">
                    {events.map((event, index) => (
                      <div className="timeline-item" key={event.title}>
                        <div className="timeline-rail">
                          <span className="timeline-dot"><Icon name={event.icon} size={14}/></span>
                          {index < events.length - 1 && <span className="timeline-line"/>}
                        </div>
                        <div className="event-content">
                          <span className="event-date">{event.date}</span>
                          <div className="event-heading"><h3>{event.title}</h3><span className="event-type">{event.kind}</span></div>
                          <p>{event.place}</p>
                          <span className="signed"><Icon name="check" size={12}/> Digitally signed</span>
                        </div>
                      </div>
                    ))}
                  </div>
                  <div className="route-map" aria-label="Map of East Africa showing shipment route">
                    <svg viewBox="0 0 420 480" className="map-art" aria-hidden="true">
                      <path className="land land-one" d="M19 26 148 12l41 50 87 13 20 58 70 54-8 72 43 72-77 129-152-9-62-72-53-14 9-92-47-64 43-75z"/>
                      <path className="land land-two" d="m300 16 99 3-5 117-75-26z"/>
                      <path className="border" d="m128 90 37 54-16 76 75 36 26 64 75 28M54 210l95 10M224 256l79-65"/>
                      <path className="route-line" d="M130 278 C157 261 165 251 185 238 S245 205 297 173"/>
                      <circle className="route-ring" cx="130" cy="278" r="10"/><circle className="route-point" cx="130" cy="278" r="5"/>
                      <circle className="route-ring" cx="185" cy="238" r="9"/><circle className="route-point" cx="185" cy="238" r="4"/>
                      <circle className="route-ring" cx="297" cy="173" r="10"/><circle className="route-point" cx="297" cy="173" r="5"/>
                    </svg>
                    <div className="map-label origin-label"><span>1</span><div><small>ORIGIN</small><strong>Nyamagabe, Rwanda</strong></div></div>
                    <div className="map-label port-label"><span>4</span><div><small>EXPORT</small><strong>Mombasa, Kenya</strong></div></div>
                    <div className="map-credit">Map data · OpenStreetMap</div>
                  </div>
                </div>
              ) : (
                <div className="genealogy">
                  <div className="genealogy-intro"><span className="section-icon"><Icon name="layers"/></span><div><h3>Product genealogy</h3><p>This lot traces to one verified origin with one recorded transformation.</p></div></div>
                  <div className="tree">
                    <div className="tree-node current"><span className="node-icon"><Icon name="box"/></span><div><small>CURRENT LOT</small><strong>CT-RW-2026-04821</strong><p>Green coffee · 2,400 kg</p></div><span className="pill verified"><Icon name="check" size={12}/> Verified</span></div>
                    <div className="tree-connector"><span>Processed at 20% yield</span></div>
                    <div className="tree-node"><span className="node-icon origin"><Icon name="map"/></span><div><small>ORIGIN LOT</small><strong>CT-RW-2026-03118</strong><p>Coffee cherry · 12,000 kg</p></div><span className="pill verified"><Icon name="check" size={12}/> Verified</span></div>
                  </div>
                </div>
              )}
            </section>

            <section className="card">
              <div className="section-heading">
                <div><span className="eyebrow">Documentation</span><h2>Certificates & evidence</h2></div>
                <span className="count-badge">3 records</span>
              </div>
              <div className="certificate-list">
                {[
                  ['Certificate of origin', 'Rwanda Development Board', 'Valid until 31 Dec 2026'],
                  ['Phytosanitary certificate', 'Rwanda Agriculture Board', 'Valid until 08 Sep 2026'],
                  ['Specialty grade analysis', 'Coffee Quality Institute', 'Issued 12 Jun 2026'],
                ].map((certificate) => (
                  <button className="certificate" key={certificate[0]}>
                    <span className="certificate-icon"><Icon name="file"/></span>
                    <span className="certificate-copy"><strong>{certificate[0]}</strong><small>{certificate[1]} · {certificate[2]}</small></span>
                    <span className="pill verified"><Icon name="check" size={12}/> Valid</span>
                    <Icon name="external" size={16}/>
                  </button>
                ))}
              </div>
            </section>
          </div>

          <aside className="side-column">
            <section className="card checks-card">
              <div className="section-heading compact"><div><span className="eyebrow">Integrity summary</span><h2>What we checked</h2></div><span className="score">7/7</span></div>
              <div className="checks-list">
                {checks.map(([title, subtitle]) => (
                  <div className="check-row" key={title}>
                    <span className="check-icon"><Icon name="check" size={14}/></span>
                    <div><strong>{title}</strong><p>{subtitle}</p></div>
                  </div>
                ))}
              </div>
              <button className="full-text-button">View all verification checks <Icon name="chevron" size={15}/></button>
            </section>

            <section className="card origin-card">
              <div className="eyebrow">Verified origin</div>
              <div className="origin-heading">
                <span className="facility-icon"><Icon name="map"/></span>
                <div><h2>Nyamagabe Highlands Farm</h2><p>Farm plot · Southern Province, Rwanda</p></div>
              </div>
              <dl>
                <div><dt>Facility status</dt><dd><span className="pill verified"><Icon name="check" size={12}/> Licensed & active</span></dd></div>
                <div><dt>Operator</dt><dd>Imena Growers Cooperative</dd></div>
                <div><dt>Coordinates</dt><dd><code>−2.4762, 29.5564</code></dd></div>
                <div><dt>Capacity check</dt><dd className="teal-text"><Icon name="check" size={13}/> Within seasonal limit</dd></div>
              </dl>
              <button className="outline-button full-width"><Icon name="map"/> View facility details</button>
            </section>

            <section className="help-card">
              <span className="help-icon"><Icon name="info"/></span>
              <div><h3>How verification works</h3><p>ChainTrack compares signed business records with tamper-evident network proofs.</p><button>Learn about the checks <Icon name="chevron" size={14}/></button></div>
            </section>

            <button className="download-button"><Icon name="download"/> Download verification report</button>
          </aside>
        </div>
      </main>

      <footer>
        <div className="page-container footer-inner">
          <Logo/>
          <p>Transparent journeys. Trusted goods.</p>
          <div><button>About</button><button>Privacy</button><button>Verification guide</button></div>
        </div>
      </footer>
    </div>
  )
}

type AppPage =
  | 'dashboard' | 'facilities' | 'register-facility' | 'issue' | 'passports'
  | 'handover' | 'transform' | 'scan' | 'certificates' | 'flags'
  | 'verification' | 'templates' | 'users' | 'reports' | 'profile'

const navGroups: { label: string; items: { id: AppPage; label: string; icon: IconName; count?: number }[] }[] = [
  { label: 'Workspace', items: [
    { id: 'dashboard', label: 'Overview', icon: 'layers' },
    { id: 'passports', label: 'Passports', icon: 'box', count: 18 },
    { id: 'handover', label: 'Handovers', icon: 'route', count: 3 },
    { id: 'facilities', label: 'Facilities', icon: 'map' },
    { id: 'transform', label: 'Transform', icon: 'link' },
  ]},
  { label: 'Trust & compliance', items: [
    { id: 'certificates', label: 'Certificates', icon: 'file', count: 2 },
    { id: 'flags', label: 'Flags', icon: 'info', count: 1 },
    { id: 'verification', label: 'Verification queue', icon: 'shield', count: 4 },
  ]},
  { label: 'Manage', items: [
    { id: 'templates', label: 'Templates', icon: 'layers' },
    { id: 'users', label: 'People & roles', icon: 'shield' },
    { id: 'reports', label: 'Reports & audit', icon: 'file' },
  ]},
]

const pageTitles: Record<AppPage, [string, string]> = {
  dashboard: ['Good morning, Aline', 'Here is what needs your attention today.'],
  facilities: ['Facility registry', 'Manage licensed origins and seasonal capacity.'],
  'register-facility': ['Register facility', 'Add a verified place of origin or processing.'],
  issue: ['Issue origin passport', 'Create a traceable identity for a new lot.'],
  passports: ['Passports', 'Track every lot your organisation currently holds.'],
  handover: ['Custody handovers', 'Send, receive and sign for goods securely.'],
  transform: ['Wash and dry cherry', 'Record how input lots become new outputs.'],
  scan: ['Scan a passport', 'Point the camera at the QR label on the bag.'],
  certificates: ['Certificates', 'Issue and manage accredited evidence.'],
  flags: ['Flags & investigations', 'Review exceptions and protect affected lots.'],
  verification: ['Organisation verification', 'Review applications and supporting documents.'],
  templates: ['Commodity templates', 'Set rules that every passport must follow.'],
  users: ['People & roles', 'Control access across your organisation.'],
  reports: ['Reports & audit', 'Review activity and export sealed records.'],
  profile: ['Your profile', 'Manage personal details and signing security.'],
}

function StatusChip({ tone = 'neutral', children }: { tone?: 'success' | 'warning' | 'danger' | 'neutral'; children: ReactNode }) {
  return <span className={`status-chip ${tone}`}>{tone === 'success' && <Icon name="check" size={12}/>} {children}</span>
}

function Meter({ value, danger = false, label }: { value: number; danger?: boolean; label: string }) {
  return <div className="meter-block"><div className="meter-label"><span>{label}</span><strong>{value}%</strong></div><div className={`meter ${danger ? 'danger' : ''}`}><i style={{ width: `${Math.min(value, 100)}%` }}/></div></div>
}

function PageHeader({ page, action, onAction }: { page: AppPage; action?: string; onAction?: () => void }) {
  const [title, subtitle] = pageTitles[page]
  return <div className="app-page-header"><div><h1>{title}</h1><p>{subtitle}</p></div>{action && <button className="primary-button" onClick={onAction}>{action.startsWith('Scan') ? <Icon name="scan"/> : <span className="plus">+</span>}{action}</button>}</div>
}

const passportRows = [
  ['CT-RW-2026-04821', 'Nyamagabe Bourbon', '2,400 kg', 'Aster Roasters', 'Ready', 'success'],
  ['CT-RW-2026-04796', 'Kivu Natural Arabica', '1,820 kg', 'Imena Growers', 'In transit', 'warning'],
  ['CT-RW-2026-04682', 'Coko Red Bourbon', '960 kg', 'Imena Growers', 'Flagged', 'danger'],
  ['CT-RW-2026-04551', 'Washed Green Coffee', '3,100 kg', 'Horizon Export', 'Ready', 'success'],
]

function DataTable({ compact = false, onOpen }: { compact?: boolean; onOpen?: () => void }) {
  return <div className="data-table-wrap"><table className="data-table"><thead><tr><th>Lot</th><th>Product</th><th>Quantity</th><th>Holder</th><th>Status</th></tr></thead><tbody>{passportRows.slice(0, compact ? 3 : 4).map((row) => <tr key={row[0]} onClick={onOpen}><td><code>{row[0]}</code></td><td><strong>{row[1]}</strong></td><td><code>{row[2]}</code></td><td>{row[3]}</td><td><StatusChip tone={row[5] as 'success' | 'warning' | 'danger'}>{row[4]}</StatusChip></td></tr>)}</tbody></table></div>
}

function Dashboard({ go }: { go: (page: AppPage) => void }) {
  return <><PageHeader page="dashboard" action="Issue passport" onAction={() => go('issue')}/>
    <div className="metric-grid">
      {[['Passports held','18','2 ready to export'],['Incoming','3','Needs your signature'],['Outgoing','1','Awaiting receiver'],['Open flags','1','Review required']].map((m,i)=><div className={`metric-card ${i===3?'alert':''}`} key={m[0]}><span>{m[0]}</span><strong>{m[1]}</strong><p>{m[2]}</p></div>)}
    </div>
    <div className="quick-actions">
      {[['Issue passport','box','issue'],['Scan QR','scan','scan'],['Hand over','route','handover'],['Register facility','map','register-facility']].map(a=><button key={a[0]} onClick={()=>go(a[2] as AppPage)}><span><Icon name={a[1] as IconName}/></span>{a[0]}<Icon name="chevron" size={15}/></button>)}
    </div>
    <div className="dashboard-grid">
      <section className="app-card"><div className="app-card-head"><div><h2>Incoming handovers</h2><p>Arriving lots that need inspection</p></div><button onClick={()=>go('handover')}>View all</button></div>
        <div className="handover-list">{[['Nyamagabe Bourbon','Horizon Coffee Exporters','2,400 kg','Today, 16:30'],['Kivu Natural Arabica','Virunga Logistics','1,820 kg','Tomorrow']].map(x=><button className="handover-row" key={x[0]} onClick={()=>go('handover')}><span className="commodity-icon"><Icon name="box"/></span><span><strong>{x[0]}</strong><small>From {x[1]}</small></span><span><code>{x[2]}</code><small>{x[3]}</small></span><Icon name="chevron" size={16}/></button>)}</div>
      </section>
      <section className="app-card capacity-panel"><div className="app-card-head"><div><h2>Facility capacity</h2><p>Current season</p></div></div><div className="capacity-content"><strong>Nyamagabe Highlands</strong><small>Farm plot · Coffee</small><Meter value={78} label="9,360 of 12,000 kg used"/><p><Icon name="check" size={14}/> You have <b>2,640 kg</b> left this season.</p><button onClick={()=>go('facilities')}>View facility</button></div></section>
    </div>
    <section className="app-card"><div className="app-card-head"><div><h2>Passports you hold</h2><p>Recently updated lots</p></div><button onClick={()=>go('passports')}>View all</button></div><DataTable compact onOpen={()=>go('passports')}/></section>
  </>
}

function Facilities({ go }: { go: (page: AppPage) => void }) {
  return <><PageHeader page="facilities" action="Register facility" onAction={()=>go('register-facility')}/><div className="facility-layout">
    <section className="app-card facility-list"><div className="search-row"><span><Icon name="scan" size={16}/><input placeholder="Search facilities"/></span><button>Filters</button></div>
      {[['Nyamagabe Highlands Farm','Farm plot · 1,200 kg left','LIC-RW-39482','Active','success'],['Kivu Belt Washing Station','Washing station · 94% used','LIC-RW-28511','Near capacity','warning'],['Coko Collection Point','Collection point · Licence expired','LIC-RW-19812','Expired','danger']].map((f,i)=><button className={`facility-row ${i===0?'selected':''}`} key={f[0]}><span className="facility-pin"><Icon name="map"/></span><span><strong>{f[0]}</strong><small>{f[1]}</small><code>{f[2]}</code></span><StatusChip tone={f[4] as 'success'|'warning'|'danger'}>{f[3]}</StatusChip></button>)}
    </section><section className="facility-map"><div className="map-grid"/><span className="map-marker m1"><Icon name="map"/></span><span className="map-marker m2"><Icon name="map"/></span><div className="selected-facility"><div><StatusChip tone="success">Active facility</StatusChip><h2>Nyamagabe Highlands Farm</h2><p>Southern Province, Rwanda</p></div><Meter value={78} label="10,800 kg of 12,000 kg"/><strong>You have 1,200 kg left this season</strong><div><button>Edit boundary</button><button className="primary-button" onClick={()=>go('issue')}>Issue passport here</button></div></div></section>
  </div></>
}

function Stepper({ steps, active }: { steps: string[]; active: number }) {
  return <div className="flow-stepper">{steps.map((step,i)=><div className={i<active?'done':i===active?'active':''} key={step}><span>{i<active?<Icon name="check" size={13}/>:i+1}</span><strong>{step}</strong></div>)}</div>
}

function RegisterFacility({ go }: { go: (page: AppPage)=>void }) {
  return <><PageHeader page="register-facility"/><Stepper steps={['Location','Details','Licence','Capacity']} active={1}/><div className="form-layout"><section className="app-card form-card"><div className="field-group"><label>Facility name</label><input defaultValue="Nyamagabe Highlands Farm"/></div><div className="field-group"><label>Facility type</label><div className="choice-grid"><button className="chosen"><Icon name="map"/>Farm plot</button><button><Icon name="layers"/>Processor</button><button><Icon name="box"/>Warehouse</button></div></div><div className="two-fields"><div className="field-group"><label>Licence number</label><input defaultValue="LIC-RW-39482"/></div><div className="field-group"><label>Licence expiry</label><input defaultValue="31 Dec 2027"/></div></div><div className="two-fields"><div className="field-group"><label>Declared capacity</label><input defaultValue="12,000"/></div><div className="field-group"><label>Period</label><select defaultValue="season"><option value="season">Per season</option></select></div></div><div className="form-actions"><button onClick={()=>go('facilities')}>Save draft</button><button className="primary-button">Continue to licence <Icon name="chevron"/></button></div></section><section className="app-card geo-panel"><div className="geo-map"><span className="boundary"/><span className="map-marker m1"><Icon name="map"/></span></div><div className="map-tools"><button className="chosen">Drop pin</button><button>Draw boundary</button></div><div className="auto-checks"><h3>Automatic checks</h3><p><Icon name="check"/> Inside registered country</p><p><Icon name="check"/> No facility overlap found</p><p><Icon name="check"/> No protected forest intersection</p></div></section></div></>
}

function IssuePassport() {
  const [quantity,setQuantity]=useState(2400)
  const exceeded=quantity>2640
  return <><PageHeader page="issue"/><Stepper steps={['Facility','Quantity','Evidence','Sign','Done']} active={1}/><div className="narrow-flow"><section className="app-card form-card"><div className="selected-summary"><span className="facility-icon"><Icon name="map"/></span><div><small>ISSUING FACILITY</small><strong>Nyamagabe Highlands Farm</strong><p>Farm plot · Active licence</p></div><button>Change</button></div><div className="field-group"><label>Commodity template</label><select><option>Rwanda specialty coffee · v3</option><option>Coltan mineral lot · v2</option></select></div><div className="field-group"><label>Quantity to issue</label><div className={`quantity-field ${exceeded?'has-error':''}`}><input type="number" value={quantity} onChange={e=>setQuantity(Number(e.target.value))}/><span>kg</span></div>{exceeded&&<small className="error-copy">Quantity exceeds remaining seasonal capacity by {quantity-2640} kg.</small>}</div><Meter value={Math.round((9360+quantity)/120)} danger={exceeded} label={`${(9360+quantity).toLocaleString()} of 12,000 kg after issue`}/>{exceeded?<div className="notice danger"><Icon name="info"/><div><strong>Capacity check failed</strong><p>Reduce the quantity or request a reviewed capacity increase.</p></div></div>:<div className="notice success"><Icon name="check"/><div><strong>Capacity check passed</strong><p>{(2640-quantity).toLocaleString()} kg will remain this season.</p></div></div>}<div className="evidence-box"><h3>Required evidence</h3><p><Icon name="check"/> Plot geolocation confirmed</p><p><Icon name="check"/> Active operating licence</p><p><span/> Harvest declaration will be added next</p></div><div className="form-actions"><button>Back</button><button className="primary-button" disabled={exceeded}>Continue to evidence <Icon name="chevron"/></button></div></section></div></>
}

function Passports() {
  const [detail,setDetail]=useState(false)
  if(detail) return <><button className="back-link" onClick={()=>setDetail(false)}>← All passports</button><PageHeader page="passports"/><section className="passport-detail-banner"><div><StatusChip tone="success">Verified passport</StatusChip><h1>Nyamagabe Bourbon</h1><code>CT-RW-2026-04821</code></div><div><small>QUANTITY</small><strong>2,400 kg</strong><small>60 × 40 kg bags</small></div><div><small>CURRENT HOLDER</small><strong>Aster Roasters Ltd.</strong><small>Received 22 Jun 2026</small></div><button className="primary-button">Hand over</button></section><div className="detail-sections"><section className="app-card"><div className="app-card-head"><div><h2>Journey</h2><p>4 signed events</p></div></div><div className="mini-timeline">{events.map(e=><div key={e.title}><span><Icon name={e.icon}/></span><div><small>{e.date}</small><strong>{e.title}</strong><p>{e.place}</p></div></div>)}</div></section><section className="app-card"><div className="app-card-head"><div><h2>Made from</h2><p>Genealogy</p></div></div><div className="lineage-node"><Icon name="box"/><div><small>ORIGIN LOT</small><code>CT-RW-2026-03118</code><p>12,000 kg coffee cherry · 20% yield</p></div></div><details><summary>Technical details</summary><p>Record fingerprints and anchoring references are available here for auditors.</p></details></section></div></>
  return <><PageHeader page="passports" action="Scan passport"/><section className="app-card"><div className="table-toolbar"><div className="tabs compact-tabs"><button className="active">Held by us <b>18</b></button><button>In transit <b>4</b></button><button>Retired</button></div><div className="search-row"><span><Icon name="scan" size={15}/><input placeholder="Search lot or product"/></span><button>Filter</button></div></div><DataTable onOpen={()=>setDetail(true)}/></section></>
}

function Handovers() {
  const [inspect,setInspect]=useState(false)
  return <><PageHeader page="handover" action="New handover"/><div className="tabs page-tabs"><button className="active">Incoming <b>3</b></button><button>Outgoing <b>1</b></button><button>Done</button></div>{inspect?<div className="narrow-flow"><Stepper steps={['Scan','Inspect','Decision','Sign']} active={1}/><section className="app-card form-card"><div className="selected-summary"><span className="commodity-icon"><Icon name="box"/></span><div><small>LOT</small><strong>Nyamagabe Bourbon</strong><code>CT-RW-2026-04821</code></div><StatusChip tone="success">Sender signed</StatusChip></div><div className="weight-compare"><div><small>SIGNED QUANTITY</small><strong>2,400.0 <i>kg</i></strong></div><span>vs</span><div><small>WEIGHT RECEIVED</small><strong>2,396.5 <i>kg</i></strong></div></div><div className="notice success"><Icon name="check"/><div><strong>Difference 3.5 kg — within tolerance</strong><p>0.15% difference · allowed tolerance 0.5%</p></div></div><div className="inspection-list"><label><input type="checkbox" defaultChecked/>Seal number matches <code>SL-995021</code></label><label><input type="checkbox" defaultChecked/>Packaging is intact</label><label><input type="checkbox"/>No signs of contamination</label></div><div className="decision-actions"><button className="danger-button">Reject & raise flag</button><button className="primary-button">Accept & sign</button></div></section></div>:<section className="app-card handover-inbox">{[['Nyamagabe Bourbon','Horizon Coffee Exporters','2,400 kg','Arrives today'],['Kivu Natural Arabica','Virunga Logistics','1,820 kg','Arrives tomorrow'],['Washed Green Coffee','Imena Growers','960 kg','Overdue']].map((h,i)=><button className="inbox-row" key={h[0]} onClick={()=>setInspect(true)}><span className="commodity-icon"><Icon name="route"/></span><span><small>FROM {h[1]}</small><strong>{h[0]}</strong><code>CT-RW-2026-0{4821-i*25}</code></span><span><strong>{h[2]}</strong><small className={i===2?'danger-text':''}>{h[3]}</small></span><StatusChip tone={i===2?'danger':'warning'}>{i===2?'Action overdue':'Inspect'}</StatusChip><Icon name="chevron"/></button>)}</section>}</>
}

function Transform() {
  const [output,setOutput]=useState(2400)
  const max=2400, exceeded=output>max
  return <><PageHeader page="transform"/><div className="transform-grid"><section className="app-card transform-column"><div className="app-card-head"><div><span className="step-number">1</span><h2>Input lots</h2><p>12,000 kg total</p></div><button>+ Add input</button></div><div className="lot-block"><Icon name="box"/><div><strong>Fresh coffee cherry</strong><code>CT-RW-2026-03118</code><small>Nyamagabe Highlands · 12,000 kg available</small></div><StatusChip tone="success">Verified</StatusChip></div></section><section className="app-card transform-column"><div className="app-card-head"><div><span className="step-number">2</span><h2>Output passports</h2><p>What was produced</p></div><button>+ Add output</button></div><div className="field-group"><label>Product name</label><input defaultValue="Washed green coffee"/></div><div className="field-group"><label>Output quantity</label><div className={`quantity-field ${exceeded?'has-error':''}`}><input type="number" value={output} onChange={e=>setOutput(Number(e.target.value))}/><span>kg</span></div></div></section></div><section className={`balance-card ${exceeded?'failed':''}`}><div><span className="balance-icon"><Icon name={exceeded?'info':'check'}/></span><div><small>MASS BALANCE</small><h2>{exceeded?'Exceeded by '+(output-max)+' kg':'Within limit · '+(max-output)+' kg headroom'}</h2><p>12,000 kg input × 20.0% maximum yield = 2,400 kg allowed output</p></div></div><div className="balance-visual"><span style={{width:`${Math.min(output/max*100,100)}%`}}/><i/></div><button className="primary-button" disabled={exceeded}>Check and sign transform</button></section></>
}

function Scanner() {
  return <><PageHeader page="scan"/><div className="scanner-shell"><div className="offline-banner">No signal. Scans work; actions will send when you reconnect.</div><div className="camera-view"><div className="scan-corners"/><div className="scan-line"/><span className="qr-art">▦</span><div><h2>Point at the label on the bag</h2><p>Works with damaged labels at close range</p></div></div><button className="manual-code">Enter lot code instead</button></div></>
}

function Certificates() {
  const [issuing,setIssuing]=useState(false)
  return <><PageHeader page="certificates" action="Issue certificate" onAction={()=>setIssuing(true)}/>{issuing?<div className="form-layout"><section className="app-card form-card"><div className="field-group"><label>Attach certificate to</label><div className="choice-grid"><button className="chosen">Passport</button><button>Facility</button></div></div><div className="field-group"><label>Passport</label><select><option>CT-RW-2026-04821 · Nyamagabe Bourbon</option></select></div><div className="two-fields"><div className="field-group"><label>Certificate type</label><select><option>Certificate of origin</option></select></div><div className="field-group"><label>Certificate number</label><input defaultValue="COO-RW-2026-8401"/></div></div><div className="two-fields"><div className="field-group"><label>Valid from</label><input defaultValue="24 Jun 2026"/></div><div className="field-group"><label>Valid until</label><input defaultValue="24 Jun 2027"/></div></div><div className="upload-zone"><Icon name="file"/><strong>Drop signed certificate here</strong><small>PDF up to 10 MB</small></div><div className="form-actions"><button onClick={()=>setIssuing(false)}>Cancel</button><button className="primary-button">Sign and issue</button></div></section><section className="app-card preview-panel"><span className="eyebrow">Live public preview</span><div className="certificate-preview"><span className="certificate-icon"><Icon name="file"/></span><StatusChip tone="success">Valid</StatusChip><h2>Certificate of origin</h2><code>COO-RW-2026-8401</code><hr/><small>ISSUED BY</small><strong>Rwanda Development Board</strong><small>VALID UNTIL</small><strong>24 June 2027</strong></div><div className="auto-checks"><h3>Checks before issuing</h3><p><Icon name="check"/> Passport is verified</p><p><Icon name="check"/> Origin licence is valid</p><p><Icon name="check"/> Your accreditation covers this type</p></div></section></div>:<section className="app-card"><div className="table-toolbar"><div className="tabs compact-tabs"><button className="active">Issued by you <b>24</b></button><button>Received <b>18</b></button><button>Revoked <b>1</b></button></div></div><div className="certificate-admin-list">{[['COO-RW-2026-8401','Certificate of origin','CT-RW-2026-04821','31 Dec 2026','Valid','success'],['PHY-RW-2026-1192','Phytosanitary','CT-RW-2026-04796','08 Sep 2026','Expiring','warning'],['GRA-RW-2025-3021','Grade analysis','CT-RW-2025-09314','12 Jan 2026','Revoked','danger']].map(c=><div key={c[0]}><span className="certificate-icon"><Icon name="file"/></span><code>{c[0]}</code><strong>{c[1]}</strong><code>{c[2]}</code><span>{c[3]}</span><StatusChip tone={c[5] as 'success'|'warning'|'danger'}>{c[4]}</StatusChip><button>•••</button></div>)}</div></section>}</>
}

function Flags() {
  return <><PageHeader page="flags"/><div className="split-queue"><section className="app-card queue-list"><div className="search-row"><span><Icon name="scan" size={15}/><input placeholder="Search flags"/></span><button>All severity</button></div>{[['Quantity discrepancy','FLG-2026-0163','Aster Roasters · Today','High'],['Licence expired','FLG-2026-0159','Coko Collection · 2d','Medium'],['Certificate expiring','FLG-2026-0148','Horizon Export · 5d','Low']].map((f,i)=><button className={i===0?'selected':''} key={f[1]}><span className={`severity-dot s${i}`}/><span><strong>{f[0]}</strong><code>{f[1]}</code><small>{f[2]}</small></span><StatusChip tone={i===0?'danger':i===1?'warning':'neutral'}>{f[3]}</StatusChip></button>)}</section><section className="app-card investigation"><div className="investigation-head"><div><StatusChip tone="danger">High severity</StatusChip><h2>Quantity discrepancy on arrival</h2><code>FLG-2026-0163</code></div><span>Day 1 · Assigned to you</span></div><div className="key-values"><div><span>Lot</span><code>CT-RW-2026-04821</code></div><div><span>Raised by</span><strong>Aster Roasters Ltd.</strong></div><div><span>What happened</span><strong>Received weight is above allowed tolerance</strong></div></div><div className="evidence-diff"><h3>Signed vs received</h3><p className="removed"><b>−</b><code>signed_quantity: 2,400.0 kg</code></p><p className="added"><b>+</b><code>received_quantity: 2,360.0 kg</code></p><small>40 kg short · 1.67% difference · tolerance 0.5%</small></div><div className="affected"><h3>Affected downstream lots</h3><div><code>CT-RW-2026-04821</code><StatusChip tone="warning">On hold</StatusChip></div></div><div className="decision-actions"><button>Request information</button><button>Clear flag</button><button className="danger-button">Confirm fraud</button></div></section></div></>
}

function VerificationQueue() {
  return <><PageHeader page="verification"/><div className="tabs page-tabs"><button className="active">Organisations <b>4</b></button><button>Certificate issuers <b>2</b></button><button>Approved</button><button>Rejected</button></div><div className="split-queue"><section className="app-card queue-list">{[['Imena Growers Cooperative','Producer · Rwanda','Waiting 3 days'],['Virunga Mineral Trading','Exporter · Rwanda','Needs changes'],['Kivu Belt Processing','Processor · DRC','Waiting 1 day']].map((o,i)=><button className={i===0?'selected':''} key={o[0]}><span className="org-avatar">{o[0].split(' ').map(x=>x[0]).slice(0,2)}</span><span><strong>{o[0]}</strong><small>{o[1]}</small><small>{o[2]}</small></span>{i===1&&<StatusChip tone="warning">Document issue</StatusChip>}</button>)}</section><section className="app-card investigation"><div className="investigation-head"><div><span className="eyebrow">Organisation application</span><h2>Imena Growers Cooperative</h2><p>Submitted 21 June 2026</p></div><StatusChip tone="warning">Reviewing</StatusChip></div><div className="key-values"><div><span>Country</span><strong>Rwanda</strong></div><div><span>Registration</span><code>RCA/2021/009481</code></div><div><span>Organisation type</span><strong>Producer cooperative</strong></div></div><div className="document-preview"><span><Icon name="file" size={28}/></span><div><strong>Registration certificate</strong><small>registration-certificate.pdf · 1.8 MB</small></div><button>Open document</button></div><div className="auto-checks"><h3>Automatic checks</h3><p><Icon name="check"/> Registry record found</p><p><Icon name="check"/> Legal name matches document</p><p><Icon name="check"/> No duplicate registration</p><p><Icon name="check"/> No sanctions match</p></div><div className="decision-actions"><button className="danger-button">Reject</button><button>Ask for changes</button><button className="primary-button">Approve organisation</button></div></section></div></>
}

function Templates() {
  return <><PageHeader page="templates" action="New template"/><div className="template-top"><div><StatusChip tone="success">v3 live</StatusChip><h2>Rwanda specialty coffee</h2><p>kg · Seasonal yield capacity</p></div><button>Preview as producer</button><button className="primary-button">Publish v4</button></div><div className="template-layout"><section className="app-card stage-editor"><div className="app-card-head"><div><h2>Stage 2 · Washing station</h2><p>Configure allowed activity and evidence</p></div></div><div className="field-group"><label>Stage name</label><input defaultValue="Wash and dry cherry"/></div><div className="field-group"><label>Allowed events</label><div className="tag-toggle">{['Receive','Transform','Split','Merge','Hand over','Export'].map((x,i)=><button className={i<3?'on':''} key={x}>{i<3&&<Icon name="check" size={11}/>} {x}</button>)}</div></div><div className="two-fields"><div className="field-group"><label>Maximum yield</label><div className="quantity-field"><input defaultValue="20.0"/><span>%</span></div></div><div className="field-group"><label>Handover tolerance</label><div className="quantity-field"><input defaultValue="0.5"/><span>%</span></div></div></div><div className="evidence-box"><h3>Required evidence</h3><label><input type="checkbox" defaultChecked/>Receiving note</label><label><input type="checkbox" defaultChecked/>Processing log</label><label><input type="checkbox"/>Quality analysis</label></div></section><section className="app-card flow-preview"><span className="eyebrow">Live flow preview</span>{['Origin farm','Washing station','Exporter','Importer'].map((s,i)=><div key={s}><span>{i+1}</span><strong>{s}</strong><small>{i===0?'Issue · Evidence':i===1?'Receive · Transform · Hand over':'Receive · Hand over'}</small></div>)}</section></div></>
}

function Users() {
  return <><PageHeader page="users" action="Invite person"/><div className="tabs page-tabs"><button className="active">People <b>12</b></button><button>Roles <b>5</b></button><button>Sign-in policy</button></div><section className="app-card"><div className="table-toolbar"><div className="search-row"><span><Icon name="scan" size={15}/><input placeholder="Search people"/></span><button>All roles</button><button>All status</button></div></div><div className="people-list">{[['AM','Aline Mukamana','aline@imena.rw','Organisation admin','2FA · PIN','Active'],['JK','Jean Kamanzi','jean@imena.rw','Facility manager','2FA · PIN','Active'],['CN','Claudine Niyonsaba','+250 78 220 9182','Operations','Password','Invited'],['EE','Eric Emery','eric@imena.rw','Viewer','2FA · PIN','Suspended']].map((u,i)=><div key={u[1]}><span className="avatar">{u[0]}</span><span><strong>{u[1]}</strong><small>{u[2]}</small></span><span>{u[3]}</span><code>{u[4]}</code><StatusChip tone={i<2?'success':i===2?'warning':'danger'}>{u[5]}</StatusChip><button>•••</button></div>)}</div></section></>
}

function Reports() {
  return <><PageHeader page="reports" action="Export CSV"/><div className="report-tiles">{[['Volume by commodity','Monthly','12,840 kg'],['Verification outcomes','Weekly','96.8% verified'],['Flags & resolutions','On demand','4 open']].map(r=><button key={r[0]}><span><Icon name="file"/></span><strong>{r[0]}</strong><small>{r[1]}</small><b>{r[2]}</b><Icon name="chevron"/></button>)}</div><section className="app-card"><div className="table-toolbar filter-toolbar"><button>01 Jun – 30 Jun 2026</button><button>All organisations</button><button>All events</button><button>All commodities</button><button className="clear-filter">Clear</button></div><div className="audit-list">{[['24 Jun 2026 · 14:32:08','Passport verified','Public verifier','CT-RW-2026-04821','success'],['24 Jun 2026 · 13:04:51','Handover accepted','Aster Roasters','HND-2026-3081','success'],['24 Jun 2026 · 11:18:29','Flag raised','System rule','FLG-2026-0163','danger'],['23 Jun 2026 · 17:46:02','Certificate issued','Rwanda Dev. Board','COO-RW-2026-8401','success']].map(a=><div className={a[4]==='danger'?'failed-row':''} key={a[0]}><code>{a[0]}</code><span className={`audit-icon ${a[4]}`}><Icon name={a[4]==='danger'?'info':'check'}/></span><strong>{a[1]}</strong><span>{a[2]}</span><code>{a[3]}</code><button>View</button></div>)}</div><div className="table-footer"><span>Showing 4 of 1,284 events</span><span><Icon name="shield" size={14}/> Every row is sealed against changes</span></div></section></>
}

function Profile() {
  return <><PageHeader page="profile"/><div className="profile-layout"><aside className="app-card profile-nav"><button className="active">Personal details</button><button>Password</button><button>Signing PIN</button><button>Two-factor security</button><button>Recovery codes</button></aside><section className="app-card form-card"><div className="profile-heading"><span className="avatar large">AM</span><div><h2>Aline Mukamana</h2><p>Organisation administrator · Imena Growers</p></div></div><div className="two-fields"><div className="field-group"><label>Full name</label><input defaultValue="Aline Mukamana"/></div><div className="field-group"><label>Email</label><input defaultValue="aline@imena.rw"/></div></div><div className="two-fields"><div className="field-group"><label>Phone</label><input defaultValue="+250 78 220 9182"/></div><div className="field-group"><label>Language</label><select><option>English</option><option>Kinyarwanda</option><option>Français</option></select></div></div><div className="security-card"><span><Icon name="shield"/></span><div><strong>Two-factor security is on</strong><p>Your account uses an authenticator app and signing PIN.</p></div><StatusChip tone="success">Protected</StatusChip></div><div className="form-actions"><button>Cancel</button><button className="primary-button">Save changes</button></div></section></div></>
}

function BusinessApp({ onPublic }: { onPublic: () => void }) {
  const [page,setPage]=useState<AppPage>('dashboard')
  const [dark,setDark]=useState(false)
  useEffect(()=>{ document.documentElement.dataset.theme=dark?'dark':'light' },[dark])
  const render = () => {
    switch(page) {
      case 'dashboard': return <Dashboard go={setPage}/>
      case 'facilities': return <Facilities go={setPage}/>
      case 'register-facility': return <RegisterFacility go={setPage}/>
      case 'issue': return <IssuePassport/>
      case 'passports': return <Passports/>
      case 'handover': return <Handovers/>
      case 'transform': return <Transform/>
      case 'scan': return <Scanner/>
      case 'certificates': return <Certificates/>
      case 'flags': return <Flags/>
      case 'verification': return <VerificationQueue/>
      case 'templates': return <Templates/>
      case 'users': return <Users/>
      case 'reports': return <Reports/>
      case 'profile': return <Profile/>
    }
  }
  return <div className="business-shell">
    <aside className="app-sidebar"><button className="sidebar-logo" onClick={onPublic}><Logo/></button><div className="org-switcher"><span className="org-avatar">IG</span><div><strong>Imena Growers</strong><small>Producer</small></div><Icon name="chevron" size={14}/></div><nav>{navGroups.map(g=><div className="nav-group" key={g.label}><label>{g.label}</label>{g.items.map(item=><button className={page===item.id?'active':''} onClick={()=>setPage(item.id)} key={item.id}><Icon name={item.icon}/><span>{item.label}</span>{item.count&&<b>{item.count}</b>}</button>)}</div>)}</nav><div className="sidebar-user"><button onClick={()=>setPage('profile')}><span className="avatar">AM</span><span><strong>Aline Mukamana</strong><small>Organisation admin</small></span><Icon name="chevron" size={14}/></button></div></aside>
    <div className="app-main"><header className="app-topbar"><div className="mobile-brand"><Logo/></div><button className="back-button" onClick={()=>setPage('dashboard')}>← <span>Back</span></button><div className="topbar-context"><span>Workspace</span><strong>{pageTitles[page][0]}</strong></div><div className="app-top-actions"><button className="icon-button" aria-label="Notifications"><Icon name="clock"/><i>3</i></button><button className="icon-button" onClick={()=>setDark(!dark)} aria-label="Theme"><Icon name={dark?'sun':'moon'}/></button><button className="outline-button" onClick={()=>setPage('scan')}><Icon name="scan"/> Scan QR</button></div></header><main className="app-content">{render()}</main></div>
    <nav className="mobile-tabs">{navGroups[0].items.slice(0,4).map(item=><button className={page===item.id?'active':''} onClick={()=>setPage(item.id)} key={item.id}><Icon name={item.icon}/><span>{item.label}</span>{item.count&&<b>{item.count}</b>}</button>)}</nav>
  </div>
}

function AuthFlow({ onDone }: { onDone: () => void }) {
  const [step,setStep]=useState<'signin'|'signup'|'org'|'documents'|'pending'>('signin')
  return <div className="auth-shell"><aside className="auth-brand"><Logo/><div><span className="eyebrow">Trusted trade infrastructure</span><h1>Every product has a story.<br/>Make it verifiable.</h1><p>Create trusted passports, sign custody handovers and protect your supply chain.</p></div><small>Secure by design · Built for regulated trade</small></aside><main className="auth-main"><div className="auth-mobile-logo"><Logo/></div>{step==='signin'&&<div className="auth-form"><span className="eyebrow">Welcome back</span><h1>Sign in to ChainTrack</h1><p>Use your verified work account.</p><div className="field-group"><label>Email</label><input placeholder="you@organisation.com"/></div><div className="field-group"><label>Password</label><input type="password" placeholder="Your password"/></div><div className="auth-options"><label><input type="checkbox"/>Keep me signed in</label><button>Forgot password?</button></div><button className="primary-button auth-submit" onClick={onDone}>Sign in securely</button><p className="auth-switch">New to ChainTrack? <button onClick={()=>setStep('signup')}>Create an account</button></p></div>}{step==='signup'&&<div className="auth-form"><span className="eyebrow">Create account · 1 of 3</span><h1>Start with your details</h1><div className="field-group"><label>Full name</label><input placeholder="As shown on your ID"/></div><div className="field-group"><label>Email or phone</label><input placeholder="you@organisation.com"/></div><div className="field-group"><label>Password</label><input type="password" placeholder="At least 12 characters"/><small>Use 12+ characters with a number and symbol.</small></div><div className="field-group"><label>Language</label><select><option>English</option><option>Kinyarwanda</option><option>Français</option></select></div><label className="consent"><input type="checkbox"/>I agree to the terms and privacy notice.</label><button className="primary-button auth-submit" onClick={()=>setStep('org')}>Create account</button></div>}{step==='org'&&<div className="auth-form wide"><span className="eyebrow">Organisation · 2 of 3</span><h1>Tell us about your organisation</h1><div className="field-group"><label>Legal organisation name</label><input placeholder="Exactly as on registration"/></div><div className="two-fields"><div className="field-group"><label>Country</label><select><option>Rwanda</option><option>Kenya</option><option>DR Congo</option></select></div><div className="field-group"><label>Registration number</label><input placeholder="RCA/2024/000000"/></div></div><div className="field-group"><label>Organisation type</label><div className="choice-grid org-types">{['Producer','Processor','Logistics','Exporter / importer','Certifier / regulator'].map((x,i)=><button className={i===0?'chosen':''} key={x}>{x}</button>)}</div></div><button className="primary-button auth-submit" onClick={()=>setStep('documents')}>Continue to documents</button></div>}{step==='documents'&&<div className="auth-form wide"><span className="eyebrow">Documents · 3 of 3</span><h1>Verify your organisation</h1><p>Upload clear, current documents. You can finish later.</p>{[['Registration certificate','registration-certificate.pdf','Uploaded'],['Operating licence','operating-licence.pdf','Uploading · 72%']].map((d,i)=><div className="upload-row" key={d[0]}><span><Icon name="file"/></span><div><strong>{d[0]}</strong><small>{d[1]}</small>{i===1&&<i><b/></i>}</div><StatusChip tone={i===0?'success':'warning'}>{d[2]}</StatusChip></div>)}<div className="upload-zone"><Icon name="file"/><strong>Add optional tax registration</strong><small>PDF, JPG or PNG up to 10 MB</small></div><p className="autosave"><Icon name="check"/> Saved 1 min ago · Finish later</p><button className="primary-button auth-submit" onClick={()=>setStep('pending')}>Submit for verification</button></div>}{step==='pending'&&<div className="auth-form wide pending"><span className="verdict-icon"><Icon name="clock"/></span><h1>We’re checking your organisation</h1><p>Most reviews finish within two business days. We’ll email you when there’s an update.</p><Stepper steps={['Account','Documents','Registry check','Approval']} active={2}/><div className="notice warning"><Icon name="info"/><div><strong>Issuing is locked for now</strong><p>You can’t issue or receive passports until verification is done.</p></div></div><button className="primary-button auth-submit" onClick={onDone}>Explore your workspace</button></div>}</main></div>
}

function App() {
  const [mode,setMode]=useState<'public'|'auth'|'business'>('public')
  if(mode==='auth') return <AuthFlow onDone={()=>setMode('business')}/>
  if(mode==='business') return <BusinessApp onPublic={()=>setMode('public')}/>
  return <PublicVerification onBusiness={()=>setMode('auth')}/>
}

export default App
