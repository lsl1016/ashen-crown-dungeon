const $ = (q) => document.querySelector(q);
const $$ = (q) => [...document.querySelectorAll(q)];
const state = { world:null, snapshot:null, runId:null, selectedClass:'warden', busy:false, sound:true, audio:null, actorRoom:null };

const roomIcons = {
  entrance:'⌂', combat:'⚔', event:'◈', treasure:'◇', npc:'♟', merchant:'¤', rest:'✦', shrine:'✧', boss:'♛',
  chapel:'✟', library:'▤', flooded:'≋', prison:'▥', garden:'❀', ossuary:'☷', forge:'⚒', observatory:'✺', banquet:'♜', secret:'◉',
  infirmary:'✚', gatehouse:'▥', aqueduct:'≈', bridge:'⌁', court:'⚖', belltower:'♢', reliquary:'◆', mausoleum:'▰',
  frontier:'⌂', ashfield:'〰', caravan:'♟', glassmarsh:'◇', crater:'✹', starwatch:'✺', windshrine:'〽', meteor:'✦', ruins:'⚑', windcamp:'♨', outergate:'┃', finalboss:'◉'
};
const typeNames = {
  entrance:'墓城入口', combat:'危险区域', event:'遗迹异象', treasure:'陪葬宝藏', npc:'幸存者营地', merchant:'无脸集市', rest:'安全余火', shrine:'禁忌祭坛', boss:'烬冠核心',
  chapel:'悼亡礼拜堂', library:'王庭档案区', flooded:'沉水城区', prison:'旧王囚区', garden:'地下王庭花园', ossuary:'千骨堂', forge:'王庭工坊', observatory:'地下观测台', banquet:'最后宴厅', secret:'隐藏王室区域',
  infirmary:'灰疫医馆', gatehouse:'折冠门楼', aqueduct:'黑水引渠', bridge:'断月桥', court:'灰烬审判庭', belltower:'无钟之塔', reliquary:'王室封藏间', mausoleum:'白石王陵',
  frontier:'雾外界碑', ashfield:'灰风平原', caravan:'停滞商队', glassmarsh:'镜砂洼地', crater:'坠星裂谷', starwatch:'反向观测台', windshrine:'无像风祠', meteor:'陨铁采掘场', ruins:'无旗骑士营', windcamp:'逐风营火', outergate:'外封印黑门', finalboss:'门后心室'
};
const attrNames = { strength:'力量', dexterity:'敏捷', perception:'感知', will:'意志' };
const kindNames = { lore:'调查', event:'触发', loot:'搜寻', npc:'对话', object:'操作', secret:'机关', rest:'休整' };
const slotNames = { weapon:'武器', armor:'护甲', trinket:'饰品' };
const rarityNames = { common:'普通', uncommon:'精良', rare:'稀有', legendary:'传说' };
const combatOnlyItems = new Set(['frost_salt','smoke_bomb','ember_flask','storm_phial']);

async function api(url, options={}) {
  const res = await fetch(url, {headers:{'Content-Type':'application/json', ...(options.headers||{})}, ...options});
  const body = await res.json().catch(()=>({}));
  if(!res.ok) throw new Error(body.error || `请求失败 ${res.status}`);
  return body;
}
function toast(msg){
  const el=$('#toast'); el.textContent=msg; el.classList.add('show');
  clearTimeout(el._t); el._t=setTimeout(()=>el.classList.remove('show'),2600);
}
function escapeHtml(s=''){ return String(s).replace(/[&<>'"]/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;',"'":'&#39;','"':'&quot;'}[c])); }
function interacted(run, room, el){ return !!run.flags[`interacted:${room.id}:${el.id}`]; }
function hasItem(run,id){ return (run.player.inventory||[]).includes(id); }
function countItem(run,id){ return (run.player.inventory||[]).filter(x=>x===id).length; }
function roman(n){ return ['0','I','II','III','IV','V','VI','VII','VIII','IX','X','XI','XII','XIII'][Math.max(0,Math.min(13,n||0))]||String(n); }

async function init(){
  state.world = await api('/api/world');
  $('#worldPremise').textContent = state.world.premise;
  renderClasses();
  $('#newGameBtn').onclick = createGame;
  $('#openSavesBtn').onclick = openSaves;
  $('#worldBtn').onclick = openWorld;
  $('#randomSeedBtn').onclick = ()=>{$('#seedInput').value=Math.floor(Math.random()*900000000)+100000000;};
  $('#saveBtn').onclick = saveGame;
  $('#journalBtn').onclick = openJournal;
  $('#codexBtn').onclick = openCodex;
  $('#talentBtn').onclick = openTalents;
  $('#soundBtn').onclick = toggleSound;
  $('#menuBtn').onclick = ()=>{ if(confirm('返回标题画面？当前 Run 已自动保存，手动存档需要点击右上角 ▣。')) location.reload(); };
  $('#closeModal').onclick = closeModal;
  $('.modal-backdrop').onclick = closeModal;
  document.addEventListener('keydown', handleHotkeys);
}

function renderClasses(){
  $('#classCards').innerHTML = state.world.classes.map(c=>`<article class="class-card ${c.id===state.selectedClass?'active':''}" data-class="${c.id}"><div class="class-top"><span class="class-ico">${c.icon}</span><div><b>${escapeHtml(c.name)}</b><small>生命 ${c.hp} · 防御 ${c.defense} · 能量 ${c.energy}</small></div></div><p>${escapeHtml(c.description)}<br><span style="color:#a08a5d">职业技：${escapeHtml(c.skillName)}</span></p></article>`).join('');
  $$('.class-card').forEach(el=>el.onclick=()=>{state.selectedClass=el.dataset.class;renderClasses();});
}

async function createGame(){
  if(state.busy)return; state.busy=true;
  try{
    const name=$('#playerName').value.trim()||'无名旅者';
    const seedRaw=$('#seedInput').value.trim();
    const seed=seedRaw?Number(seedRaw):Math.floor(Date.now()/1000);
    const snap=await api('/api/runs',{method:'POST',body:JSON.stringify({name,class:state.selectedClass,seed})});
    startSnapshot(snap); toast('V0.4 世界已建立 · 第一幕：沉眠墓城'); sfx('start');
  }catch(e){toast(e.message)}finally{state.busy=false}
}
function startSnapshot(snap){
  state.snapshot=snap; state.runId=snap.run.id;
  $('#startScreen').classList.add('hidden'); $('#gameShell').classList.remove('hidden');
  render();
}
async function refresh(){ if(!state.runId)return; state.snapshot=await api(`/api/runs/${state.runId}`); render(); }

function render(){
  const {run,items}=state.snapshot, p=run.player, room=run.rooms[run.currentRoomId];
  if(!p.equipment) p.equipment={}; if(!p.talents)p.talents={};
  const cls=state.world.classes.find(c=>c.id===p.class)||state.world.classes[0];
  $('#playerNameLabel').textContent=p.name; $('#classNameLabel').textContent=cls.name; $('#classIcon').textContent=cls.icon; $('#levelBadge').textContent=`Lv.${p.level}`;
  $('#seedLabel').textContent=run.seed; $('#bellLabel').textContent=`${roman(run.clock?.bell||1)} / XIII`; $('#roomBreadcrumb').textContent=`${room.zone||'墓城'} / ${room.name}`; $('#turnLabel').textContent=`TURN ${run.turn}`;
  $('#talentPointBadge').textContent=p.talentPoints||0; $('#talentPointBadge').classList.toggle('empty',!(p.talentPoints>0));
  setBar('#hpBar',p.hp,p.maxHp); setBar('#energyBar',p.energy,p.maxEnergy); const xpNeed=p.level*40; setBar('#xpBar',p.xp,xpNeed);
  $('#hpText').textContent=`${p.hp} / ${p.maxHp}`; $('#energyText').textContent=`${p.energy} / ${p.maxEnergy}`; $('#xpText').textContent=`${p.xp} / ${xpNeed}`; $('#goldLabel').textContent=p.gold;
  $('#statsGrid').innerHTML=[['力量',p.attributes.strength],['敏捷',p.attributes.dexterity],['感知',p.attributes.perception],['意志',p.attributes.will],['防御',p.defense],['天赋点',p.talentPoints||0]].map(([k,v])=>`<div class="stat"><span>${k}</span><b>${v}</b></div>`).join('');
  renderEquipment(items,p); renderInventory(items,p); renderScene(room,run); renderStory(run); renderContext(run,room); renderActions(run,room); renderMap(run); renderNearby(run); renderQuests(run); renderProverb(run); renderWorldPulse(run);
}
function setBar(sel,v,max){ const el=$(sel); if(el)el.style.width=`${Math.max(0,Math.min(100,max?((v/max)*100):0))}%`; }

function renderEquipment(items,p){
  const slots=['weapon','armor','trinket'];
  $('#equipmentCard').innerHTML=slots.map(slot=>{
    const id=p.equipment?.[slot],it=id?items[id]:null;
    if(!it)return `<div class="equip-slot empty"><span>${slotNames[slot]}</span><b>空</b></div>`;
    const bonus=[]; if(it.power)bonus.push(`威力 +${it.power}`);if(it.defense)bonus.push(`防御 +${it.defense}`);if(it.maxHp)bonus.push(`生命 +${it.maxHp}`);if(it.maxEnergy)bonus.push(`能量 +${it.maxEnergy}`);
    return `<div class="equip-slot rarity-${it.rarity||'common'}" title="${escapeHtml(it.description)}"><span>${slotNames[slot]}</span><div><i>${it.icon}</i><b>${escapeHtml(it.name)}</b></div><small>${escapeHtml(bonus.join(' · ')||it.special||'已装备')}</small></div>`;
  }).join('');
}

function renderInventory(items,p){
  const counts={}; p.inventory.forEach(id=>counts[id]=(counts[id]||0)+1); $('#inventoryCount').textContent=p.inventory.length;
  const equipped=new Set(Object.values(p.equipment||{})); const inCombat=!!state.snapshot.run.combat;
  $('#inventoryList').innerHTML=Object.entries(counts).map(([id,n])=>{
    const it=items[id]||{name:id,icon:'?',description:'未知物品',type:'unknown',power:0}; const isEquipped=equipped.has(id);
    let action='';
    if(it.slot) action=isEquipped?'<span class="item-count">已装备</span>':`<button class="mini-btn" data-item="${id}" data-item-action="equip">装备</button>`;
    else if(it.type==='consumable'){
      const combatOnly=combatOnlyItems.has(id);
      const canUse=combatOnly?inCombat:!inCombat;
      if(canUse) action=`<button class="mini-btn" data-item="${id}" data-item-action="use">${combatOnly?'战斗用':'使用'}</button>`;
    }
    return `<div class="item ${isEquipped?'equipped':''} rarity-${it.rarity||'common'}" title="${escapeHtml(it.description)}"><div class="item-icon">${it.icon}</div><div class="item-body"><b>${escapeHtml(it.name)} <em>${rarityNames[it.rarity]||''}</em></b><small>${escapeHtml(it.description)}</small></div><div class="item-side">${n>1?`<span class="item-count">×${n}</span>`:''}${action}</div></div>`;
  }).join('')||'<p class="muted" style="font-size:9px">背包是空的。</p>';
  $$('[data-item-action]').forEach(b=>b.onclick=(ev)=>{ev.stopPropagation();itemAction(b.dataset.item,b.dataset.itemAction);});
}

function elementState(run,room,el){
  if(el.hiddenUnlessFlag && !run.flags[el.hiddenUnlessFlag])return {visible:false,locked:true,reason:'尚未发现'};
  const done=el.oneShot&&interacted(run,room,el); if(done)return {visible:true,done:true,locked:true,reason:'已完成'};
  if(el.requiresItem&&!hasItem(run,el.requiresItem))return {visible:true,locked:true,reason:`需要 ${state.snapshot.items[el.requiresItem]?.name||el.requiresItem}`};
  if(el.requiresFlag&&!run.flags[el.requiresFlag])return {visible:true,locked:true,reason:'需要先改变附近环境或取得线索'};
  return {visible:true,done:false,locked:false,reason:''};
}

function renderScene(room,run){
  const scene=$('#scene'); scene.className=`scene scene-${room.scene} threat-${Math.min(4,run.clock?.threat||0)}`;
  renderPlayerActor(room,run);
  $('#zoneLabel').textContent=room.zone||'墓城'; $('#roomTypeLabel').textContent=run.activeEvent?run.activeEvent.title:(typeNames[room.type]||room.type);
  $('#roomTitle').textContent=room.name; $('#roomDescription').textContent=run.activeEvent?run.activeEvent.description:room.description;
  const eventCard=$('#eventCard');
  if(run.activeEvent){ eventCard.classList.remove('hidden'); $('#eventTitle').textContent=run.activeEvent.title; $('#eventDescription').textContent=run.activeEvent.description; } else eventCard.classList.add('hidden');

  const canInteract=!run.combat&&!run.activeEvent&&!run.gameOver; let available=0, lockedCount=0;
  $('#hotspotLayer').innerHTML=(room.elements||[]).map(el=>{
    const st=elementState(run,room,el); if(!st.visible)return ''; if(!st.done)available++; if(st.locked&&!st.done)lockedCount++;
    const clickable=canInteract&&!st.locked; const check=el.check?` · ${attrNames[el.check.attribute]} DC${el.check.dc}`:'';
    return `<div class="hotspot ${escapeHtml(el.kind)} ${st.done?'done':''} ${st.locked?'locked':''}" style="left:${el.x}%;top:${el.y}%" ${clickable?`data-interact="${el.id}"`:''} title="${escapeHtml(st.reason||el.description)}${check}"><span class="hotspot-pin">${st.done?'✓':st.locked?'🔒':escapeHtml(el.icon)}</span><span class="hotspot-label">${escapeHtml(el.label)}${el.check?` <i>DC${el.check.dc}</i>`:''}</span></div>`;
  }).join('');
  $('#interactionCounter').textContent=run.combat?'战斗中':run.activeEvent?'事件处理中':`${available} 个目标${lockedCount?` · ${lockedCount} 个受条件限制`:''}`;
  $$('[data-interact]').forEach(el=>el.onclick=()=>interact(el.dataset.interact));

  const es=$('#enemyStage');
  if(run.combat){
    es.classList.remove('hidden'); const enemy=state.snapshot.enemies[run.combat.enemyId], intent=run.combat.intent||{};
    $('#enemyIcon').textContent=enemy?.icon||'☠'; $('#enemyName').textContent=run.combat.enemyName; $('#enemyDescription').textContent=enemy?.description||''; $('#enemyArchetype').textContent=`${enemy?.archetype||'敌人'} · 弱点 ${enemy?.weakness||'未知'}`;
    $('#enemyHpText').textContent=`${Math.max(0,run.combat.enemyHp)} / ${run.combat.enemyMaxHp}`; setBar('#enemyHpBar',Math.max(0,run.combat.enemyHp),run.combat.enemyMaxHp);
    $('#intentIcon').textContent=intent.icon||'⚔'; $('#intentLabel').textContent=intent.label||'未知意图'; $('#intentDescription').textContent=`${intent.description||''}${intent.telegraph?` · ${intent.telegraph}`:''}`;
    $('#enemyStatuses').innerHTML=(run.combat.enemyStatuses||[]).map(s=>`<span class="status-chip enemy">${escapeHtml(s.name)} ${s.rounds}</span>`).join('') || '<span class="status-empty">无异常状态</span>';
    const phase=$('#bossPhase'); if(enemy?.boss){phase.classList.remove('hidden');phase.textContent=`PHASE ${roman(run.combat.bossPhase||1)}`}else phase.classList.add('hidden');
  } else es.classList.add('hidden');
}

function renderStory(run){
  const logs=run.log.slice(-24);
  $('#storyFeed').innerHTML=logs.map((l,i)=>`<div class="story-line ${l.type} ${i===logs.length-1?'latest':''}"><small>${String(l.turn).padStart(2,'0')}</small>${escapeHtml(l.message)}</div>`).join('');
  const feed=$('#storyFeed'); feed.scrollTop=feed.scrollHeight;
}
function renderContext(run,room){
  const box=$('#contextBar');
  if(run.combat){
    const intent=run.combat.intent||{}; const ps=(run.combat.playerStatuses||[]).map(s=>`${s.name} ${s.rounds}`).join(' · ');
    box.innerHTML=`<span class="context-chip intent">${intent.icon||'⚔'} 下一步：${escapeHtml(intent.label||'未知')}</span><span class="context-chip">${escapeHtml(intent.telegraph||'观察敌人动作决定应对方式')}</span>${ps?`<span class="context-chip debuff">自身状态：${escapeHtml(ps)}</span>`:''}`; return;
  }
  if(run.activeEvent){box.innerHTML=`<span class="context-chip">◈ ${escapeHtml(run.activeEvent.title)}</span><span class="context-chip">选择会永久写入本次世界状态</span>`;return;}
  const els=(room.elements||[]).map(el=>({el,st:elementState(run,room,el)})).filter(x=>x.st.visible);
  box.innerHTML=els.length?els.map(({el,st})=>`<button class="context-chip ${st.done?'done':''} ${st.locked&&!st.done?'locked':''}" ${st.done||st.locked?'disabled':`data-context-interact="${el.id}"`}>${st.done?'✓':st.locked?'🔒':'◌'} ${escapeHtml(el.label)}${st.locked&&!st.done?` · ${escapeHtml(st.reason)}`:''}</button>`).join(''):'<span class="context-chip">这里没有明显可调查目标</span>';
  $$('[data-context-interact]').forEach(b=>b.onclick=()=>interact(b.dataset.contextInteract));
}

function adjacentRooms(run){
  const ids=[]; for(const e of run.edges){if(e.from===run.currentRoomId)ids.push(e.to);else if(e.to===run.currentRoomId)ids.push(e.from)}
  return [...new Set(ids)].map(id=>run.rooms[id]).filter(Boolean);
}
function renderActions(run,room){
  const box=$('#actionArea');
  if(run.gameOver){box.innerHTML=`<button class="action-btn primary" onclick="location.reload()">${run.victory?'完成冒险 · 返回标题':'重新开始'}</button><span class="action-hint">${run.victory?'墓城的命运已经改变。':'你的旅程停在了这里。'}</span>`;return;}
  if(run.combat){
    const cls=state.world.classes.find(c=>c.id===run.player.class), skill=cls?.skillName||'职业技', cd=run.combat.cooldowns?.skill||0;
    const potion=countItem(run,'healing_draught');
    box.innerHTML=`<button class="action-btn primary" data-combat="attack" data-hotkey="1"><kbd>1</kbd>⚔ 普通攻击</button><button class="action-btn" data-combat="skill" data-hotkey="2" ${cd>0||run.player.energy<3?'disabled':''} title="${escapeHtml(cls?.skillDescription||'')}"><kbd>2</kbd>✦ ${escapeHtml(skill)} ${cd>0?`· 冷却 ${cd}`:'· 3 能量'}</button><button class="action-btn" data-combat="guard" data-hotkey="3"><kbd>3</kbd>⛨ 防御</button><button class="action-btn" data-combat="potion" data-hotkey="4" ${potion<=0?'disabled':''}><kbd>4</kbd>🧪 药剂 ×${potion}</button><button class="action-btn danger" data-combat="flee" data-hotkey="5"><kbd>5</kbd>↶ 撤退</button><span class="action-hint">ROUND ${run.combat.round+1} · 先看敌人意图再行动</span>`;
    $$('[data-combat]').forEach(b=>b.onclick=()=>combatAction(b.dataset.combat)); return;
  }
  if(run.activeEvent){
    box.innerHTML=run.activeEvent.choices.map((c,i)=>`<button class="action-btn ${c.check?'primary':''}" data-choice="${c.id}" data-hotkey="${i+1}"><kbd>${i+1}</kbd>${escapeHtml(c.text)}${c.check?` · ${attrNames[c.check.attribute]} DC${c.check.dc}`:''}</button>`).join('')+`<span class="action-hint">${escapeHtml(run.activeEvent.title)}</span>`;
    $$('[data-choice]').forEach(b=>b.onclick=()=>chooseEvent(b.dataset.choice)); return;
  }
  const els=(room.elements||[]).map(el=>({el,st:elementState(run,room,el)})).filter(x=>x.st.visible&&!x.st.done);
  const nearby=adjacentRooms(run).filter(r=>r.discovered);
  let hotkey=1, html='';
  for(const {el,st} of els.slice(0,4)){
    if(st.locked){html+=`<button class="action-btn locked" disabled>🔒 ${escapeHtml(el.label)} · ${escapeHtml(st.reason)}</button>`;continue;}
    html+=`<button class="action-btn" data-action-interact="${el.id}" data-hotkey="${hotkey}"><kbd>${hotkey++}</kbd>${kindNames[el.kind]||'调查'} · ${escapeHtml(el.label)}${el.check?` · DC${el.check.dc}`:''}</button>`;
  }
  for(const r of nearby){
    if(r.locked){html+=`<button class="action-btn locked" disabled>🔒 ${escapeHtml(r.visited?r.name:(typeNames[r.type]||'未知道路'))}</button>`;continue;}
    html+=`<button class="action-btn travel" data-move="${r.id}" data-hotkey="${hotkey}"><kbd>${hotkey++}</kbd>→ 前往 ${escapeHtml(r.visited?r.name:(typeNames[r.type]||'未知地点'))}</button>`;
  }
  html+=`<span class="action-hint">条件机关会明确显示需要的钥匙、线索或检定</span>`; box.innerHTML=html;
  $$('[data-action-interact]').forEach(b=>b.onclick=()=>interact(b.dataset.actionInteract)); $$('[data-move]').forEach(b=>b.onclick=()=>moveTo(b.dataset.move));
}

async function chooseEvent(id){ sfx('interact'); await perform(`/api/runs/${state.runId}/action`,{choiceId:id},true,'event'); }
async function combatAction(action){ if(action==='skill')actorEmote('cast');else if(action==='guard')actorEmote('guard'); sfx(action==='skill'?'skill':action==='guard'?'guard':'attack'); await perform(`/api/runs/${state.runId}/combat`,{action},true,'combat'); }
async function itemAction(itemId,action){ sfx(action==='equip'?'equip':'item'); await perform(`/api/runs/${state.runId}/item`,{itemId,action},false,'item'); }
async function interact(elementId){
  const run=state.snapshot.run, room=run.rooms[run.currentRoomId], el=(room.elements||[]).find(x=>x.id===elementId);
  if(el) await approachElement(el);
  sfx('interact');
  await perform(`/api/runs/${state.runId}/interact`,{elementId},true,'interact');
}
async function moveTo(roomId){ const run=state.snapshot.run;if(run.combat||run.activeEvent||run.gameOver||state.busy)return;sfx('move');await perform(`/api/runs/${state.runId}/move`,{roomId},false,'move'); }

async function perform(url,payload,showDice,kind){
  if(state.busy)return; state.busy=true;
  try{
    const old=state.snapshot?.run;
    const oldHP=old?.player?.hp, oldEnergy=old?.player?.energy, oldEnemyHP=old?.combat?.enemyHp, oldPhase=old?.combat?.bossPhase;
    const oldCombat=old?.combat?{...old.combat}:null, oldRoom=old?.rooms?.[old?.currentRoomId], oldAct2=!!old?.flags?.act2_unlocked;
    const snap=await api(url,{method:'POST',body:JSON.stringify(payload)});
    if(showDice&&snap.run.lastRoll) await diceFlash(snap.run.lastRoll);
    state.snapshot=snap; render();
    const now=snap.run, nowRoom=now.rooms[now.currentRoomId];
    if(oldEnemyHP!=null){
      const nextEnemyHP=now.combat?.enemyHp ?? 0, delta=Math.max(0,oldEnemyHP-nextEnemyHP);
      if(delta>0){pulseClass('#scene','scene-hit',500);spawnFloat('enemy',`-${delta}`,now.lastRoll?.critical?'crit':'damage');spawnVfx(kind==='combat'?'slash':'burst','enemy');sfx(now.lastRoll?.critical?'crit':'hit');}
    }
    if(oldHP!=null&&now.player.hp<oldHP){const d=oldHP-now.player.hp;pulseClass('#scene','screen-shake',360);actorEmote('hit');spawnFloat('player',`-${d}`,'damage');sfx('hurt');}
    if(oldHP!=null&&now.player.hp>oldHP){spawnFloat('player',`+${now.player.hp-oldHP}`,'heal');sfx('heal');}
    if(oldEnergy!=null&&now.player.energy>oldEnergy&&kind!=='move'){spawnFloat('player',`+${now.player.energy-oldEnergy} EN`,'energy');}
    if(!oldCombat&&now.combat){pulseClass('#scene','combat-start',520);if(state.snapshot.enemies[now.combat.enemyId]?.boss){await showCinematic(now.combat.enemyId==='gate_heart'?'FINAL ENCOUNTER':'BOSS ENCOUNTER',now.combat.enemyName,state.snapshot.enemies[now.combat.enemyId]?.description||'', 'boss',1100);sfx('boss');}}
    if(oldPhase&&now.combat?.bossPhase>oldPhase){pulseClass('#scene','phase-shift',900);await showCinematic(`PHASE ${roman(now.combat.bossPhase)}`,now.combat.enemyName,'敌人的战斗模式发生了变化。','boss',850);sfx('phase');}
    if(!oldAct2&&now.flags.act2_unlocked){pulseClass('#scene','region-shift',850);await showCinematic('ACT II','雾外荒原','赫里昂倒下后，墓城背后的石门第一次打开。真正的封印在灰风尽头。','',1500);sfx('act');}
    if(kind==='move'){
      $('#scene').animate([{opacity:.32,filter:'blur(6px)'},{opacity:1,filter:'blur(0)'}],{duration:430,easing:'ease-out'});
      if(oldRoom&&nowRoom&&oldRoom.zone!==nowRoom.zone){showCinematic('REGION',nowRoom.zone,nowRoom.description,'',900);}
    }
    if(!old?.victory&&now.victory){pulseClass('#scene','final-pulse',1200);await showCinematic('THE END','封印重新闭合','灰风越过荒原。维尔的名字重新回到世界。','final',1800);sfx('victory');}
  }catch(e){toast(e.message);sfx('error')}finally{state.busy=false}
}
function pulseClass(sel,cl,ms){const el=$(sel);if(!el)return;el.classList.add(cl);setTimeout(()=>el.classList.remove(cl),ms)}

async function diceFlash(result){
  const o=$('#diceOverlay'); o.classList.remove('hidden','fail','critical'); if(!result.success)o.classList.add('fail');if(result.critical)o.classList.add('critical');
  $('#diceResultLabel').textContent=result.label||'检定';$('#diceValue').textContent='?';$('#diceText').textContent='命运正在转动……';
  let n=0;const t=setInterval(()=>{$('#diceValue').textContent=1+Math.floor(Math.random()*20);n++;if(n>10)clearInterval(t)},48);
  await new Promise(r=>setTimeout(r,570));clearInterval(t);$('#diceValue').textContent=result.roll;$('#diceText').textContent=`${result.roll} + ${result.bonus} = ${result.total}${result.dc?` / DC ${result.dc}`:''} · ${result.success?'成功':'失败'}`;
  await new Promise(r=>setTimeout(r,680));o.classList.add('hidden');
}

function renderMap(run){
  const svg=$('#mapSvg'),rooms=Object.values(run.rooms),sx=92,sy=102,ox=30,oy=52;let html='';const adj=new Set(adjacentRooms(run).map(r=>r.id));
  const maxX=Math.max(...rooms.map(r=>r.x||0)); svg.setAttribute('viewBox',`-35 -30 ${Math.max(900,ox+maxX*sx+120)} 610`);
  const dividerX=ox+9.2*sx; html+=`<line class="map-region-divider" x1="${dividerX}" y1="10" x2="${dividerX}" y2="555"/><text class="map-region-label" x="55" y="22">ACT I · 沉眠墓城</text><text class="map-region-label act2" x="${dividerX+30}" y="22">ACT II · 雾外荒原</text>`;
  for(const e of run.edges){const a=run.rooms[e.from],b=run.rooms[e.to];if(!a||!b)continue;const visible=a.discovered&&b.discovered&&!a.locked&&!b.locked;html+=`<line class="map-edge ${visible?'':'hidden-edge'}" x1="${ox+a.x*sx}" y1="${oy+a.y*sy}" x2="${ox+b.x*sx}" y2="${oy+b.y*sy}"/>`;}
  for(const r of rooms){
    const x=ox+r.x*sx,y=oy+r.y*sy,unknown=!r.discovered,reachable=adj.has(r.id)&&r.discovered&&!r.locked&&!run.combat&&!run.activeEvent&&!run.gameOver;
    const act2=(r.x||0)>=10;const cl=['map-node',unknown?'unknown':'discovered',r.id===run.currentRoomId?'current':'',r.resolved?'resolved':'',reachable?'reachable':'',r.locked&&r.discovered?'locked':'',act2?'act2':''].join(' ');
    const icon=unknown?'?':r.locked?'⌧':(roomIcons[r.type]||'•'),label=unknown?'未发现':(r.visited?r.name:(typeNames[r.type]||'未知地点'));
    html+=`<g class="${cl}" ${reachable?`data-map-move="${r.id}"`:''} transform="translate(${x},${y})"><circle r="20"></circle><text y="-1">${icon}</text><text class="node-label" y="32">${escapeHtml(label.length>8?label.slice(0,8)+'…':label)}</text></g>`;
  }
  svg.innerHTML=html;$('#mapProgress').textContent=`${rooms.filter(r=>r.discovered).length} / ${rooms.length}`;$$('[data-map-move]').forEach(n=>n.onclick=()=>moveTo(n.dataset.mapMove));
}
function renderNearby(run){
  const rooms=adjacentRooms(run).filter(r=>r.discovered);
  $('#nearbyList').innerHTML=rooms.map(r=>`<div class="nearby-row ${r.locked?'locked':''}" ${r.locked?'':`data-nearby="${r.id}"`}><span class="nearby-icon">${r.locked?'⌧':roomIcons[r.type]||'•'}</span><div><b>${escapeHtml(r.visited?r.name:(typeNames[r.type]||'未知地点'))}</b><small>${escapeHtml(r.zone||'墓城')} · ${r.locked?'条件封锁':r.resolved?'已处理':r.visited?'仍有危险':'未踏足'}</small></div><em>${r.locked?'封锁':'前往 ›'}</em></div>`).join('')||'<p class="muted" style="font-size:9px">没有已发现道路。</p>';
  $$('[data-nearby]').forEach(n=>n.onclick=()=>moveTo(n.dataset.nearby));
}
function renderQuests(run){
  const qs=Object.values(run.quests);$('#questList').innerHTML=qs.map(q=>`<div class="quest ${q.status==='completed'?'completed':''}"><b>${q.status==='completed'?'✓ ':''}${escapeHtml(q.title)}</b><p>${escapeHtml(q.description)}</p><small>${q.status==='completed'?'已完成':`${q.progress} / ${q.goal}`}</small></div>`).join('');
}
function renderWorldPulse(run){
  const c=run.clock||{bell:1,threat:0,lastChange:'第一声钟已经结束。'};$('#worldThreat').textContent=`威胁 ${c.threat||0} · 第 ${c.bell||1} 钟`;$('#worldChange').textContent=c.lastChange||'墓城暂时保持沉默。';
}
function renderProverb(run){
  let t='“王冠不是王权。它是门闩。”';if(run.flags.opened_outer_gate)t='“门不是为了阻止我们出去，而是阻止它进来。”';else if(run.flags.act2_unlocked)t='“墓城只是一枚钉子。灰风之外，还有真正的门。”';else if(run.flags.knows_last_price)t='“第十三种代价不是死亡，而是遗忘为何不肯死。”';else if(run.flags.heard_true_name)t='“真名不会打开门。真名会让守门的人醒来。”';else if(run.flags.saw_false_crown)t='“你看见的王冠，只是封印希望你看见的形状。”';$('#proverbText').textContent=t;
}

async function saveGame(){if(!state.runId)return;try{const run=state.snapshot.run;const meta=await api(`/api/runs/${state.runId}/save`,{method:'POST',body:JSON.stringify({name:`${run.player.name} · ${run.rooms[run.currentRoomId].name}`})});toast(`已保存：${meta.name}`)}catch(e){toast(e.message)}}
async function openSaves(){
  try{const saves=await api('/api/saves');showModal(`<h2>读取存档</h2><div class="save-list">${saves.length?saves.map(s=>`<div class="save-row"><div><b>${escapeHtml(s.name)}</b><p>${escapeHtml(s.playerName)} · Lv.${s.level} · ${escapeHtml(s.currentRoom)}<br>${new Date(s.createdAt).toLocaleString()}</p></div><button class="btn primary" data-load="${s.id}">读取</button></div>`).join(''):'<p class="muted">还没有手动存档。</p>'}</div>`);$$('[data-load]').forEach(b=>b.onclick=async()=>{try{const snap=await api(`/api/saves/${b.dataset.load}/load`,{method:'POST'});closeModal();startSnapshot(snap)}catch(e){toast(e.message)}})}catch(e){toast(e.message)}
}
function openJournal(){if(!state.snapshot)return;const run=state.snapshot.run;showModal(`<h2>冒险日志</h2><div class="log-list">${run.log.slice().reverse().map(l=>`<div><b>#${l.turn}</b> <span style="color:#777">[${escapeHtml(l.type)}]</span> ${escapeHtml(l.message)}</div>`).join('')}</div>`)}
function openCodex(){if(!state.snapshot)return;const lore=state.snapshot.run.lore||[];showModal(`<h2>发现档案 <small style="font-size:10px;color:#696f76">${lore.length} 条</small></h2><div class="codex-list">${lore.length?lore.map(x=>`<article class="codex-item"><b>${escapeHtml(x.title)}</b><p>${escapeHtml(x.text)}</p></article>`).join(''):'<p class="muted">还没有发现任何档案。调查碑文、病历、判决、笔记与遗物可以解锁。</p>'}</div>`)}
function openWorld(){const w=state.world;showModal(`<h2>${escapeHtml(w.title)}</h2><div class="world-info"><p>${escapeHtml(w.premise)}</p><div class="world-grid"><section class="world-box"><h3>时代</h3><p>${escapeHtml(w.era)}</p></section><section class="world-box"><h3>已知势力</h3><ul>${w.factions.map(x=>`<li>${escapeHtml(x)}</li>`).join('')}</ul></section></div><p class="muted">V0.4 把赫里昂改为第一幕 Boss。击败他会打开雾外荒原，世界地图扩展到 44 个地点；最终结局发生在第二幕黑门之后。</p></div>`)}
function openTalents(){
  if(!state.snapshot)return;const run=state.snapshot.run,p=run.player,cls=state.world.classes.find(c=>c.id===p.class);const talents=(state.world.talents||[]).filter(t=>t.class===p.class);
  showModal(`<h2>${escapeHtml(cls?.name||'职业')} · 天赋 <small style="font-size:10px;color:#c5a45d">可用 ${p.talentPoints||0} 点</small></h2><div class="talent-grid">${talents.map(t=>{const rank=p.talents?.[t.id]||0,learned=rank>=t.maxRank;return `<article class="talent-card ${learned?'learned':''}"><span class="talent-icon">${t.icon}</span><div><b>${escapeHtml(t.name)}</b><p>${escapeHtml(t.description)}</p><small>${learned?'已学习':`${rank} / ${t.maxRank}`}</small></div>${learned?'':`<button class="mini-btn" data-talent="${t.id}" ${(p.talentPoints||0)<=0?'disabled':''}>学习</button>`}</article>`}).join('')}</div><p class="muted" style="font-size:9px">每次升级获得 1 点天赋点。天赋会直接改变战斗结算，不只是文字说明。</p>`);
  $$('[data-talent]').forEach(b=>b.onclick=()=>learnTalent(b.dataset.talent));
}
async function learnTalent(id){
  if(state.busy)return;state.busy=true;try{state.snapshot=await api(`/api/runs/${state.runId}/talent`,{method:'POST',body:JSON.stringify({talentId:id})});render();openTalents();toast('天赋已学习')}catch(e){toast(e.message)}finally{state.busy=false}
}
function showModal(html){$('#modalBody').innerHTML=html;$('#modal').classList.remove('hidden')}
function closeModal(){$('#modal').classList.add('hidden')}
function handleHotkeys(e){if(e.key==='Escape'&&!$('#modal').classList.contains('hidden')){closeModal();return}if(!state.snapshot||state.busy||!$('#modal').classList.contains('hidden'))return;if(/^[1-9]$/.test(e.key)){const target=document.querySelector(`[data-hotkey="${e.key}"]`);if(target&&!target.disabled){e.preventDefault();target.click()}}}


function renderPlayerActor(room,run){
  const actor=$('#playerActor'), icon=$('#playerActorIcon'); if(!actor||!icon)return;
  const cls=state.world.classes.find(c=>c.id===run.player.class); icon.textContent=cls?.icon||'◆';
  if(state.actorRoom!==room.id){state.actorRoom=room.id;setActorPosition(run.combat?18:16,run.combat?74:76,false);}
  actor.classList.toggle('in-combat',!!run.combat);
}
function setActorPosition(x,y,animate=true){const a=$('#playerActor');if(!a)return;if(!animate)a.style.transition='none';a.style.left=`${Math.max(8,Math.min(88,x))}%`;a.style.top=`${Math.max(24,Math.min(82,y))}%`;if(!animate){requestAnimationFrame(()=>a.style.transition='');}}
async function approachElement(el){
  const actor=$('#playerActor');if(!actor)return;actor.classList.add('moving');setActorPosition(Math.max(12,Math.min(82,(el.x||50)-8)),Math.max(35,Math.min(78,(el.y||55)+10)),true);await new Promise(r=>setTimeout(r,380));actor.classList.remove('moving');
}
function actorEmote(kind){const a=$('#playerActor');if(!a)return;a.classList.remove('cast','hit','guard');void a.offsetWidth;a.classList.add(kind);setTimeout(()=>a.classList.remove(kind),650);}
function spawnFloat(target,text,kind='damage'){
  const layer=$('#vfxLayer');if(!layer)return;const el=document.createElement('div');el.className=`float-number ${kind}`;el.textContent=text;
  if(target==='enemy'){el.style.left='77%';el.style.top='43%';}else{const a=$('#playerActor');el.style.left=a?.style.left||'18%';el.style.top=a?.style.top||'73%';}
  layer.appendChild(el);setTimeout(()=>el.remove(),1150);
}
function spawnVfx(type,target='enemy'){
  const layer=$('#vfxLayer');if(!layer)return;const el=document.createElement('div');el.className=type==='burst'?'vfx-burst':type==='guard'?'vfx-guard':'vfx-slash';
  el.style.left=target==='enemy'?'68%':'16%';el.style.top=target==='enemy'?'38%':'58%';layer.appendChild(el);setTimeout(()=>el.remove(),750);
}
async function showCinematic(kicker,title,text,kind='',duration=1000){
  const o=$('#cinematicOverlay');if(!o)return;o.className=`cinematic-overlay show ${kind}`;$('#cinematicKicker').textContent=kicker;$('#cinematicTitle').textContent=title;$('#cinematicText').textContent=text||'';o.classList.remove('hidden');await new Promise(r=>setTimeout(r,duration));o.classList.add('hidden');o.classList.remove('show','boss','final');
}
function toggleSound(){state.sound=!state.sound;$('#soundBtn').textContent=`声音 ${state.sound?'ON':'OFF'}`;if(state.sound){ensureAudio();sfx('interact');}}
function ensureAudio(){
  if(state.audio)return state.audio;const Ctx=window.AudioContext||window.webkitAudioContext;if(!Ctx)return null;state.audio=new Ctx();return state.audio;
}
function sfx(kind){
  if(!state.sound)return;const ctx=ensureAudio();if(!ctx)return;if(ctx.state==='suspended')ctx.resume().catch(()=>{});
  const presets={attack:[180,95,.07],hit:[90,55,.08],crit:[520,180,.12],hurt:[120,60,.11],skill:[330,760,.16],guard:[240,300,.09],heal:[360,560,.13],item:[420,480,.08],equip:[260,390,.08],move:[100,130,.06],interact:[300,360,.05],boss:[70,42,.28],phase:[190,80,.22],act:[220,440,.24],victory:[330,660,.32],start:[220,330,.18],error:[150,100,.08]};const [a,b,d]=presets[kind]||presets.interact;
  const now=ctx.currentTime,o=ctx.createOscillator(),g=ctx.createGain();o.type=kind==='boss'||kind==='hurt'?'sawtooth':'sine';o.frequency.setValueAtTime(a,now);o.frequency.exponentialRampToValueAtTime(Math.max(35,b),now+d);g.gain.setValueAtTime(.0001,now);g.gain.exponentialRampToValueAtTime(kind==='boss'?.07:.035,now+.01);g.gain.exponentialRampToValueAtTime(.0001,now+d);o.connect(g);g.connect(ctx.destination);o.start(now);o.stop(now+d+.02);
}

init().catch(e=>{console.error(e);toast('初始化失败：'+e.message)});
