const $ = (id) => document.getElementById(id);
const state = { content:{}, kinds:[], overrides:new Set(), kind:'', selectedId:'', runs:[], runId:'', snapshot:null, selectedRoom:'' };

async function api(url, options={}) {
  const res = await fetch(url,{headers:{'Content-Type':'application/json',...(options.headers||{})},...options});
  let body={}; try{ body=await res.json(); }catch{}
  if(!res.ok) throw new Error(body.error || `HTTP ${res.status}`);
  return body;
}
function toast(msg, error=false){ const el=$('toast'); el.textContent=msg; el.className=`toast show${error?' error':''}`; clearTimeout(toast.t); toast.t=setTimeout(()=>el.className='toast',2600); }
function esc(v){ return String(v??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])); }
function pretty(v){ return JSON.stringify(v,null,2); }
function nameOf(v,id){ return v?.name || v?.title || v?.label || v?.npcId || id; }

async function boot(){
  bind();
  try{
    const [catalog, overrides, runs] = await Promise.all([api('/api/editor/content'),api('/api/editor/overrides'),api('/api/editor/runs')]);
    state.content=catalog.content||{}; state.kinds=catalog.kinds||Object.keys(state.content); state.overrides=new Set((overrides||[]).map(x=>`${x.kind}:${x.id}`)); state.runs=runs||[];
    renderKinds(); renderRuns();
    if(state.kinds.length) selectKind(state.kinds.includes('worldEvents')?'worldEvents':state.kinds[0]);
    if(state.runs.length){ state.runId=state.runs[0].id; $('runSelect').value=state.runId; await loadRun(); } else renderNoRun();
  }catch(err){ toast(err.message,true); }
}
function bind(){
  $('kindSelect').addEventListener('change',e=>selectKind(e.target.value));
  $('contentSearch').addEventListener('input',renderContentList);
  $('saveContentBtn').addEventListener('click',saveContent);
  $('newContentBtn').addEventListener('click',newContent);
  $('duplicateBtn').addEventListener('click',duplicateContent);
  $('exportBtn').addEventListener('click',exportContent);
  $('contentJson').addEventListener('input',renderContentPreview);
  $('runSelect').addEventListener('change',async e=>{state.runId=e.target.value;await loadRun();});
  $('refreshRunBtn').addEventListener('click',loadRun);
  $('advanceBellBtn').addEventListener('click',()=>worldAction({action:'advance_bell'},'世界时间已推进'));
  $('triggerEventBtn').addEventListener('click',()=>worldAction({action:'trigger_event',eventId:$('eventSelect').value},'世界事件已触发'));
  $('moveNpcBtn').addEventListener('click',()=>worldAction({action:'move_npc',npcId:$('npcSelect').value,roomId:$('npcRoomSelect').value},'NPC 已移动'));
  $('setRegionBtn').addEventListener('click',()=>worldAction({action:'set_region',region:$('regionSelect').value,value:$('regionValue').value},'区域状态已写入'));
  $('setFlagBtn').addEventListener('click',()=>worldAction({action:'set_flag',flag:$('flagName').value,boolValue:$('flagValue').value==='true'},'Flag 已设置'));
  $('roomSelect').addEventListener('change',e=>selectRoom(e.target.value));
  $('newRoomBtn').addEventListener('click',newRoom);
  $('saveRoomBtn').addEventListener('click',saveRoom);
}
function renderKinds(){ $('kindSelect').innerHTML=state.kinds.map(k=>`<option value="${esc(k)}">${esc(k)}</option>`).join(''); }
function selectKind(kind){ state.kind=kind; $('kindSelect').value=kind; state.selectedId=''; renderContentList(); const ids=Object.keys(state.content[kind]||{}).sort(); if(ids.length) selectContent(ids[0]); else newContent(); }
function renderContentList(){
  const q=$('contentSearch').value.trim().toLowerCase(), rows=state.content[state.kind]||{};
  const ids=Object.keys(rows).filter(id=>{const v=rows[id]; return !q || `${id} ${nameOf(v,id)} ${v?.description||''}`.toLowerCase().includes(q)}).sort();
  $('contentStats').textContent=`${Object.keys(rows).length} 条定义 · ${ids.length} 条可见`;
  $('contentList').innerHTML=ids.map(id=>{const v=rows[id], over=state.overrides.has(`${state.kind}:${id}`);return `<div class="content-row ${id===state.selectedId?'active':''}" data-id="${esc(id)}"><b>${esc(nameOf(v,id))}${over?' · ◈':''}</b><small>${esc(id)}</small></div>`}).join('') || '<div class="empty">该分类暂无定义</div>';
  [...$('contentList').querySelectorAll('.content-row')].forEach(el=>el.onclick=()=>selectContent(el.dataset.id));
}
function selectContent(id){ state.selectedId=id; const v=state.content[state.kind]?.[id]; if(!v)return; $('contentId').value=id; $('contentJson').value=pretty(v); $('overrideBadge').classList.toggle('hidden',!state.overrides.has(`${state.kind}:${id}`)); renderContentList(); renderContentPreview(); }
function defaultTemplate(kind,id){
  const templates={
    items:{id,name:'新物品',type:'consumable',description:'GM 创建的物品。',power:0,value:10,icon:'◇'},
    enemies:{id,name:'新敌人',description:'GM 创建的敌人。',hp:30,attack:5,defense:11,damageMin:3,damageMax:6,xp:20,goldMin:2,goldMax:8,icon:'◆',boss:false},
    skills:{id,class:'warden',name:'新技能',icon:'✦',cost:2,cooldown:2,minDistance:1,maxDistance:3,hitAttribute:'strength',hitBonus:1,description:'数据驱动技能。',effects:[{type:'damage',value:4,dice:6}],tags:['custom']},
    worldEvents:{id,title:'新世界事件',region:'外墓区',triggerBell:3,escalateBell:6,severity:1,overlay:'plague',description:'一个由 GM 创建的世界事件。'},
    questGraphs:{id,title:'新任务图',nodes:[{id:'start',label:'开始',kind:'start'},{id:'ending',label:'结局',kind:'ending'}],edges:[{from:'start',to:'ending',label:'选择'}]},
    npcs:{id,name:'新 NPC',title:'旅人',faction:'无所属',description:'GM 创建的人物。',portrait:'/assets/portraits/npc_iven.svg'},
    npcSchedules:{npcId:id,entries:[{fromBell:1,toBell:13,roomId:'room_01',activity:'等待世界变化'}]}
  }; return templates[kind] || {id};
}
function newContent(){ const id=`custom_${state.kind}_${Date.now().toString(36)}`; state.selectedId=''; $('contentId').value=id; $('contentJson').value=pretty(defaultTemplate(state.kind,id)); $('overrideBadge').classList.add('hidden'); renderContentPreview(); }
function duplicateContent(){ let val; try{val=JSON.parse($('contentJson').value);}catch{return toast('当前 JSON 无效',true)} const id=`${$('contentId').value||'custom'}_copy`; val.id=id; if('npcId' in val)val.npcId=id; state.selectedId=''; $('contentId').value=id; $('contentJson').value=pretty(val); renderContentPreview(); }
async function saveContent(){
  const id=$('contentId').value.trim(); if(!id)return toast('ID 不能为空',true); let value; try{value=JSON.parse($('contentJson').value);}catch(err){return toast('JSON 解析失败：'+err.message,true)}
  try{ await api('/api/editor/content',{method:'POST',body:JSON.stringify({kind:state.kind,id,value})}); value.id=id; if(state.kind==='npcSchedules')value.npcId=id; state.content[state.kind] ||= {}; state.content[state.kind][id]=value; state.selectedId=id; state.overrides.add(`${state.kind}:${id}`); $('overrideBadge').classList.remove('hidden'); renderContentList(); toast('内容覆盖已持久化，重启仍会加载'); }
  catch(err){toast(err.message,true)}
}
function exportContent(){ const blob=new Blob([pretty(state.content[state.kind]||{})],{type:'application/json'});const a=document.createElement('a');a.href=URL.createObjectURL(blob);a.download=`ashen-crown-${state.kind}.json`;a.click();URL.revokeObjectURL(a.href); }
function renderContentPreview(){
  let v; try{v=JSON.parse($('contentJson').value);}catch{ $('contentPreview').innerHTML='<span style="color:#d8877e">JSON 尚未合法</span>'; return; }
  if(state.kind==='questGraphs'){
    const nodes=v.nodes||[], edges=v.edges||[]; let html='<b>任务图预览</b><div class="graph-preview">';
    if(!nodes.length) html+='暂无节点'; else { nodes.forEach((n,i)=>{html+=`<span class="graph-node ${n.kind==='ending'?'ending':''}">${esc(n.label||n.id)}</span>`; if(i<nodes.length-1)html+='<span class="graph-arrow">→</span>';}); }
    html+='</div>'; if(edges.length)html+=`<small>${edges.length} 条边；保存后 World API 与未来 AI 会读取同一份定义。</small>`; $('contentPreview').innerHTML=html;
  } else if(state.kind==='skills') $('contentPreview').innerHTML=`<b>${esc(v.icon||'✦')} ${esc(v.name||'技能')}</b> · Cost ${esc(v.cost)} · CD ${esc(v.cooldown)} · Range ${esc(v.minDistance)}-${esc(v.maxDistance)}<br><small>${esc((v.effects||[]).map(x=>x.type).join(' → '))}</small>`;
  else $('contentPreview').innerHTML=`<b>${esc(nameOf(v,$('contentId').value))}</b><br><small>${esc(v.description||v.title||'结构化内容定义')}</small>`;
}
function renderRuns(){ $('runSelect').innerHTML=state.runs.length?state.runs.map(r=>`<option value="${esc(r.id)}">${esc(r.playerName)} · Lv.${r.level} · Bell ${r.bell} · ${esc(r.currentRoom)}</option>`).join(''):'<option value="">暂无 Run</option>'; }
function renderNoRun(){ $('noRun').classList.remove('hidden'); $('runWorkspace').classList.add('hidden'); }
async function loadRun(){ if(!state.runId)return renderNoRun(); try{state.snapshot=await api(`/api/runs/${encodeURIComponent(state.runId)}`); $('noRun').classList.add('hidden');$('runWorkspace').classList.remove('hidden');renderRun();}catch(err){toast(err.message,true)} }
function renderRun(){ const r=state.snapshot.run, rooms=state.snapshot.rooms||[]; if(!r)return;
  const room=r.rooms?.[r.currentRoomId]; $('runSummary').innerHTML=[['旅者',`${r.player.name} · Lv.${r.player.level}`],['位置',room?.name||r.currentRoomId],['世界钟',`${r.clock.bell} / XIII`],['威胁',r.clock.threat],['Turn',r.turn]].map(x=>`<div class="summary-cell"><small>${x[0]}</small><b>${esc(x[1])}</b></div>`).join('');
  const worldDefs=state.content.worldEvents||{}; $('eventSelect').innerHTML=Object.keys(worldDefs).sort().map(id=>`<option value="${esc(id)}">${esc(worldDefs[id].title||id)}</option>`).join('');
  const events=Object.values(r.worldEvents||{}); $('activeEvents').innerHTML=events.length?events.map(e=>`<div class="mini-row ${e.status==='active'?'event-live':'event-resolved'}"><span>${esc(e.title)} · ${esc(e.stage)}</span><small>${esc(e.region)} / S${e.severity}</small></div>`).join(''):'<div class="mini-row"><small>当前没有已调度事件</small></div>';
  const npcIds=Object.keys(state.content.npcs||{}).sort(); $('npcSelect').innerHTML=npcIds.map(id=>`<option value="${esc(id)}">${esc(nameOf(state.content.npcs[id],id))}</option>`).join(''); const roomOptions=rooms.map(x=>`<option value="${esc(x.id)}">${esc(x.name)} (${esc(x.id)})</option>`).join(''); $('npcRoomSelect').innerHTML='<option value="@schedule">恢复自动日程</option>'+roomOptions;
  $('npcWorldList').innerHTML=npcIds.map(id=>{const n=state.content.npcs[id]||{},w=r.npcWorld?.[id]||{}, rr=r.rooms?.[w.location||r.npcLocations?.[id]];return `<div class="mini-row"><span>${esc(n.name||id)} · ${esc(w.activity||'')}</span><small>${esc(rr?.name||w.location||'?')} / ${esc(w.mood||'')}</small></div>`}).join('');
  const regions=r.regionStates||{}; $('regionSelect').innerHTML=Object.keys(regions).sort().map(x=>`<option>${esc(x)}</option>`).join(''); $('regionList').innerHTML=Object.entries(regions).map(([k,v])=>`<div class="mini-row"><span>${esc(k)}</span><small>${esc(v)}</small></div>`).join(''); if($('regionSelect').value)$('regionValue').value=regions[$('regionSelect').value]||''; $('regionSelect').onchange=()=>{$('regionValue').value=regions[$('regionSelect').value]||''};
  const flags=Object.entries(r.flags||{}).filter(x=>x[1]); $('flagList').innerHTML=flags.slice(0,60).map(([k])=>`<span class="flag-chip on">${esc(k)}</span>`).join('') || '<small>尚无 true Flag</small>';
  renderRoomSelectors(rooms); if(!state.selectedRoom || !r.rooms?.[state.selectedRoom]) state.selectedRoom=r.currentRoomId; $('roomSelect').value=state.selectedRoom; selectRoom(state.selectedRoom,false);
}
function renderRoomSelectors(rooms){ const opts=rooms.map(x=>`<option value="${esc(x.id)}">${esc(x.name)} · ${esc(x.zone)}</option>`).join(''); $('roomSelect').innerHTML=opts; $('connectRoomSelect').innerHTML='<option value="">不新增连接</option>'+opts; }
function selectRoom(id, update=true){ const r=state.snapshot?.run;if(!r)return; const room=r.rooms?.[id];if(!room)return; state.selectedRoom=id;if(update)$('roomSelect').value=id;$('roomJson').value=pretty(room); const ss=r.sceneStates?.[id]; $('roomSceneState').innerHTML=ss?`<b>${esc(ss.label)}</b><br><small>${esc(ss.kind)} · ${esc(ss.description||'')}</small>`:'<small>没有 SceneState</small>'; renderMiniMap(); }
function newRoom(){ const r=state.snapshot?.run;if(!r)return;const id=`gm_room_${Date.now().toString(36)}`;state.selectedRoom='';$('roomJson').value=pretty({id,name:'GM 新地点',type:'event',scene:'hall',description:'由 V0.7 GM Editor 创建的地点。',zone:'自定义区域',x:500,y:250,discovered:true,visited:false,resolved:false,locked:false,elements:[]});$('roomSelect').value='';$('roomSceneState').innerHTML='<small>保存后生成场景状态</small>'; }
async function saveRoom(){ if(!state.runId)return;let room;try{room=JSON.parse($('roomJson').value)}catch(err){return toast('房间 JSON 无效：'+err.message,true)};try{state.snapshot=await api(`/api/editor/runs/${encodeURIComponent(state.runId)}/room`,{method:'POST',body:JSON.stringify({room,connectTo:$('connectRoomSelect').value})});state.selectedRoom=room.id;renderRun();toast('地点已写入当前 Run');}catch(err){toast(err.message,true)} }
async function worldAction(payload,msg){ if(!state.runId)return toast('没有 Run',true);try{state.snapshot=await api(`/api/editor/runs/${encodeURIComponent(state.runId)}/world`,{method:'POST',body:JSON.stringify(payload)});renderRun();toast(msg);}catch(err){toast(err.message,true)} }
function renderMiniMap(){ const r=state.snapshot?.run;if(!r)return;const rooms=Object.values(r.rooms||{}); if(!rooms.length)return; const xs=rooms.map(x=>x.x),ys=rooms.map(x=>x.y),minX=Math.min(...xs),maxX=Math.max(...xs),minY=Math.min(...ys),maxY=Math.max(...ys),sx=x=>40+(x-minX)/Math.max(1,maxX-minX)*540,sy=y=>35+(y-minY)/Math.max(1,maxY-minY)*270; let html=''; for(const e of r.edges||[]){const a=r.rooms[e.from],b=r.rooms[e.to];if(a&&b)html+=`<line class="map-edge" x1="${sx(a.x)}" y1="${sy(a.y)}" x2="${sx(b.x)}" y2="${sy(b.y)}"/>`;} for(const room of rooms){const cls=`map-node${room.id===r.currentRoomId?' current':''}${room.id===state.selectedRoom?' selected':''}`;html+=`<circle class="${cls}" cx="${sx(room.x)}" cy="${sy(room.y)}" r="${room.id===state.selectedRoom?7:5}"><title>${esc(room.name)}</title></circle>`;} $('miniMap').innerHTML=html; }

document.addEventListener('DOMContentLoaded',boot);
