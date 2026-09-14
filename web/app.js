const $ = (q) => document.querySelector(q);
const $$ = (q) => [...document.querySelectorAll(q)];
const state = { world:null, snapshot:null, runId:null, selectedClass:'warden', busy:false, sound:true, audio:null, actorRoom:null, actorMode:'idle', actorFrame:1, actorUntil:0, shopTab:'buy' };

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
const kindNames = { lore:'调查', event:'触发', loot:'搜寻', npc:'对话', dialogue:'对话', object:'操作', secret:'机关', rest:'休整' };
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
  $('#talentBtn').onclick = openGrowth;
  $('#soundBtn').onclick = toggleSound;
  $('#menuBtn').onclick = ()=>{ if(confirm('返回标题画面？当前 Run 已自动保存，手动存档需要点击右上角 ▣。')) location.reload(); };
  $('#closeModal').onclick = closeModal;
  $('.modal-backdrop').onclick = closeModal;
  document.addEventListener('keydown', handleHotkeys);
  setInterval(tickActorAnimation, 190);
  $('#scene').addEventListener('mousemove', sceneParallax);
  $('#scene').addEventListener('mouseleave', ()=>{const sc=$('#scene');['--parallax-x','--parallax-mid','--parallax-front'].forEach(v=>sc.style.removeProperty(v));});
}

function renderClasses(){
  $('#classCards').innerHTML = state.world.classes.map(c=>`<article class="class-card ${c.id===state.selectedClass?'active':''}" data-class="${c.id}"><div class="class-preview"><img src="${actorFramePath(c.id,'idle',1)}" alt="${escapeHtml(c.name)}"></div><div class="class-top"><span class="class-ico">${c.icon}</span><div><b>${escapeHtml(c.name)}</b><small>生命 ${c.hp} · 防御 ${c.defense} · 能量 ${c.energy}</small></div></div><p>${escapeHtml(c.description)}<br><span style="color:#a08a5d">职业技：${escapeHtml(c.skillName)}</span></p></article>`).join('');
  $$('.class-card').forEach(el=>el.onclick=()=>{state.selectedClass=el.dataset.class;renderClasses();});
}

async function createGame(){
  if(state.busy)return; state.busy=true;
  try{
    const name=$('#playerName').value.trim()||'无名旅者';
    const seedRaw=$('#seedInput').value.trim();
    const seed=seedRaw?Number(seedRaw):Math.floor(Date.now()/1000);
    const snap=await api('/api/runs',{method:'POST',body:JSON.stringify({name,class:state.selectedClass,seed})});
    startSnapshot(snap); toast('V0.7 Living World 已建立 · 世界事件调度已启用'); sfx('start');
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
  $('#playerNameLabel').textContent=p.name; $('#classNameLabel').textContent=cls.name; $('#classPortrait').src=actorFramePath(p.class,'idle',1); $('#levelBadge').textContent=`Lv.${p.level}`;
  $('#seedLabel').textContent=run.seed; $('#bellLabel').textContent=`${roman(run.clock?.bell||1)} / XIII`; $('#roomBreadcrumb').textContent=`${room.zone||'墓城'} / ${room.name}`; $('#turnLabel').textContent=`TURN ${run.turn}`;
  const growthPoints=(p.talentPoints||0)+(p.attributePoints||0)+(p.masteryPoints||0); $('#talentPointBadge').textContent=growthPoints; $('#talentPointBadge').classList.toggle('empty',!(growthPoints>0));
  setBar('#hpBar',p.hp,p.maxHp); setBar('#energyBar',p.energy,p.maxEnergy); const xpNeed=p.level*40; setBar('#xpBar',p.xp,xpNeed);
  $('#hpText').textContent=`${p.hp} / ${p.maxHp}`; $('#energyText').textContent=`${p.energy} / ${p.maxEnergy}`; $('#xpText').textContent=`${p.xp} / ${xpNeed}`; $('#goldLabel').textContent=p.gold;
  $('#statsGrid').innerHTML=[['力量',p.attributes.strength],['敏捷',p.attributes.dexterity],['感知',p.attributes.perception],['意志',p.attributes.will],['防御',p.defense],['成长点',(p.attributePoints||0)+(p.masteryPoints||0)+(p.talentPoints||0)]].map(([k,v])=>`<div class="stat"><span>${k}</span><b>${v}</b></div>`).join('');
  renderEquipment(items,p); renderInventory(items,p); renderScene(room,run); renderStory(run); renderContext(run,room); renderActions(run,room); renderMap(run); renderNearby(run); renderQuests(run); renderProverb(run); renderWorldPulse(run); renderWorldEvents(run);
}
function setBar(sel,v,max){ const el=$(sel); if(el)el.style.width=`${Math.max(0,Math.min(100,max?((v/max)*100):0))}%`; }

function renderEquipment(items,p){
  const slots=['weapon','armor','trinket'];
  $('#equipmentCard').innerHTML=slots.map(slot=>{
    const id=p.equipment?.[slot],it=id?items[id]:null;
    if(!it)return `<div class="equip-slot empty"><span>${slotNames[slot]}</span><b>空</b></div>`;
    const aff=(state.snapshot.run.itemAffixes?.[id]||[]); const bonus=[]; if(it.power)bonus.push(`威力 +${it.power}`);if(it.defense)bonus.push(`防御 +${it.defense}`);if(it.maxHp)bonus.push(`生命 +${it.maxHp}`);if(it.maxEnergy)bonus.push(`能量 +${it.maxEnergy}`);
    const affName=aff.length?`${aff.map(a=>a.name).join(' / ')} · `:'';
    return `<div class="equip-slot rarity-${it.rarity||'common'} ${aff.length?'affixed':''}" title="${escapeHtml(it.description)}"><span>${slotNames[slot]}</span><div><i>${it.icon}</i><b>${escapeHtml(affName+it.name)}</b></div><small>${escapeHtml(bonus.join(' · ')||it.special||'已装备')}${aff.length?` · ${escapeHtml(aff.map(a=>a.description).join('；'))}`:''}</small></div>`;
  }).join('');
}

function renderInventory(items,p){
  const counts={}; p.inventory.forEach(id=>counts[id]=(counts[id]||0)+1); $('#inventoryCount').textContent=p.inventory.length;
  const equipped=new Set(Object.values(p.equipment||{})); const inCombat=!!state.snapshot.run.combat;
  $('#inventoryList').innerHTML=Object.entries(counts).map(([id,n])=>{
    const it=items[id]||{name:id,icon:'?',description:'未知物品',type:'unknown',power:0}; const isEquipped=equipped.has(id); const aff=(state.snapshot.run.itemAffixes?.[id]||[]);
    let action='';
    if(it.slot) action=isEquipped?'<span class="item-count">已装备</span>':`<button class="mini-btn" data-item="${id}" data-item-action="equip">装备</button>`;
    else if(it.type==='consumable'){
      const combatOnly=combatOnlyItems.has(id);
      const canUse=combatOnly?inCombat:!inCombat;
      if(canUse) action=`<button class="mini-btn" data-item="${id}" data-item-action="use">${combatOnly?'战斗用':'使用'}</button>`;
    }
    return `<div class="item ${isEquipped?'equipped':''} ${aff.length?'affixed':''} rarity-${it.rarity||'common'}" title="${escapeHtml(it.description)}"><div class="item-icon">${it.icon}</div><div class="item-body"><b>${aff.length?`<i class="affix-name">${escapeHtml(aff.map(a=>a.name).join(' / '))}</i> `:''}${escapeHtml(it.name)} <em>${rarityNames[it.rarity]||''}</em></b><small>${escapeHtml(it.description)}${aff.length?` · ${escapeHtml(aff.map(a=>a.description).join('；'))}`:''}</small></div><div class="item-side">${n>1?`<span class="item-count">×${n}</span>`:''}${action}</div></div>`;
  }).join('')||'<p class="muted" style="font-size:9px">背包是空的。</p>';
  $$('[data-item-action]').forEach(b=>b.onclick=(ev)=>{ev.stopPropagation();itemAction(b.dataset.item,b.dataset.itemAction);});
}

function elementState(run,room,el){
  if(el.action==='dialogue' && run.npcLocations?.[el.target] && run.npcLocations[el.target]!==room.id)return {visible:false,locked:true,reason:'NPC 已经离开这里'};
  if(el.hiddenUnlessFlag && !run.flags[el.hiddenUnlessFlag])return {visible:false,locked:true,reason:'尚未发现'};
  const done=el.oneShot&&interacted(run,room,el); if(done)return {visible:true,done:true,locked:true,reason:'已完成'};
  if(el.requiresItem&&!hasItem(run,el.requiresItem))return {visible:true,locked:true,reason:`需要 ${state.snapshot.items[el.requiresItem]?.name||el.requiresItem}`};
  if(el.requiresFlag&&!run.flags[el.requiresFlag])return {visible:true,locked:true,reason:'需要先改变附近环境或取得线索'};
  return {visible:true,done:false,locked:false,reason:''};
}

function renderScene(room,run){
  const scene=$('#scene'), sceneState=run.sceneStates?.[room.id]||{}; scene.className=`scene scene-${room.scene} threat-${Math.min(4,run.clock?.threat||0)} zone-${room.id>='room_33'?'waste':'tomb'} world-${sceneState.kind||'stable'}`;
  renderEnvironment(room,run); renderPlayerActor(room,run); renderSceneNPCs(room,run); renderDialoguePanel(run); renderShopPanel(run);
  $('#zoneLabel').textContent=room.zone||'墓城'; $('#regionStateLabel').textContent=sceneState.label?`${sceneState.label} · ${run.regionStates?.[room.zone]||'状态未知'}`:(run.regionStates?.[room.zone]||'状态未知'); $('#roomTypeLabel').textContent=run.activeEvent?run.activeEvent.title:(typeNames[room.type]||room.type);
  $('#roomTitle').textContent=room.name; $('#roomDescription').textContent=run.activeEvent?run.activeEvent.description:room.description;
  const eventCard=$('#eventCard');
  if(run.activeEvent){ eventCard.classList.remove('hidden'); $('#eventTitle').textContent=run.activeEvent.title; $('#eventDescription').textContent=run.activeEvent.description; } else eventCard.classList.add('hidden');

  const canInteract=!run.combat&&!run.activeEvent&&!run.activeDialogue&&!run.activeShop&&!run.gameOver; let available=0, lockedCount=0;
  $('#hotspotLayer').innerHTML=(room.elements||[]).map(el=>{
    const st=elementState(run,room,el); if(!st.visible)return ''; if(!st.done)available++; if(st.locked&&!st.done)lockedCount++;
    if(el.action==='dialogue') return '';
    const clickable=canInteract&&!st.locked; const check=el.check?` · ${attrNames[el.check.attribute]} DC${el.check.dc}`:'';
    return `<div class="hotspot ${escapeHtml(el.kind)} ${st.done?'done':''} ${st.locked?'locked':''}" style="left:${el.x}%;top:${el.y}%" ${clickable?`data-interact="${el.id}"`:''} title="${escapeHtml(st.reason||el.description)}${check}"><span class="hotspot-pin">${st.done?'✓':st.locked?'🔒':escapeHtml(el.icon)}</span><span class="hotspot-label">${escapeHtml(el.label)}${el.check?` <i>DC${el.check.dc}</i>`:''}</span></div>`;
  }).join('');
  $('#interactionCounter').textContent=run.combat?'战斗中':run.activeEvent?'事件处理中':run.activeDialogue?'对话中':run.activeShop?'交易中':`${available} 个目标${lockedCount?` · ${lockedCount} 个受条件限制`:''}`;
  $$('[data-interact]').forEach(el=>el.onclick=()=>interact(el.dataset.interact));

  const es=$('#enemyStage');
  if(run.combat){
    es.classList.remove('hidden'); const enemy=state.snapshot.enemies[run.combat.enemyId], intent=run.combat.intent||{};
    const portrait=$('#enemyPortrait'); portrait.src=enemy?.portrait||'/assets/portraits/enemy_skeleton.svg'; portrait.className=`enemy-portrait intent-${escapeHtml(intent.kind||'attack')} ${enemy?.boss?'boss':''}`;
    $('#enemyName').textContent=run.combat.enemyName; $('#enemyDescription').textContent=enemy?.description||''; $('#enemyArchetype').textContent=`${enemy?.archetype||'敌人'} · 弱点 ${enemy?.weakness||'未知'}`;
    const dist=Math.max(1,Math.min(3,run.combat.distance||2)), dn=['','近距','中距','远距'][dist]; const dbox=$('#combatDistance'); dbox.querySelectorAll('i').forEach((x,i)=>x.classList.toggle('on',i<dist)); dbox.querySelector('b').textContent=dn;
    $('#enemyHpText').textContent=`${Math.max(0,run.combat.enemyHp)} / ${run.combat.enemyMaxHp}`; setBar('#enemyHpBar',Math.max(0,run.combat.enemyHp),run.combat.enemyMaxHp);
    $('#intentIcon').textContent=intent.icon||'⚔'; $('#intentLabel').textContent=intent.label||'未知意图'; $('#intentDescription').textContent=`${intent.description||''}${intent.telegraph?` · ${intent.telegraph}`:''}`;
    $('#enemyStatuses').innerHTML=(run.combat.enemyStatuses||[]).map(s=>`<span class="status-chip enemy">${escapeHtml(s.name)} ${s.rounds}</span>`).join('') || '<span class="status-empty">无异常状态</span>';
    const terrainNames={open:'开阔战场',broken_cover:'断柱掩体',blackwater:'黑水带',cold_forge:'冷炉裂口',ashstorm:'灰暴战场'}; $('#terrainLabel').textContent=terrainNames[run.combat.terrain]||run.combat.terrain||'开阔战场'; $('#terrainHint').textContent=run.combat.terrainHint||'没有额外地形修正。'; $('#terrainCard').classList.toggle('hazard',!!(run.combat.hazards||[]).length);
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
  if(run.activeDialogue){const d=currentDialogue(run);box.innerHTML=d?`<span class="context-chip npc">♟ ${escapeHtml(d.npc.name)} · ${escapeHtml(d.npc.faction)}</span><span class="context-chip">关系 ${relationshipText(run.npcRelations?.[d.npc.id]||0)} ${run.npcRelations?.[d.npc.id]||0}</span>`:'<span class="context-chip">对话中</span>';return;}
  if(run.activeShop){const shop=state.world.shops?.[run.activeShop];box.innerHTML=`<span class="context-chip shop">¤ ${escapeHtml(shop?.name||'交易')}</span><span class="context-chip">古金币 ${run.player.gold}</span><span class="context-chip">买卖会立即写入存档状态</span>`;return;}
  if(run.combat){
    const intent=run.combat.intent||{}; const ps=(run.combat.playerStatuses||[]).map(s=>`${s.name} ${s.rounds}`).join(' · ');
    const dn=['','近距','中距','远距'][Math.max(1,Math.min(3,run.combat.distance||2))];
    const hazards=(run.combat.hazards||[]).filter(h=>h.distance===run.combat.distance); box.innerHTML=`<span class="context-chip distance">◎ ${dn}</span><span class="context-chip intent">${intent.icon||'⚔'} 下一步：${escapeHtml(intent.label||'未知')}</span><span class="context-chip">${escapeHtml(run.combat.terrainHint||intent.telegraph||'观察敌人动作决定应对方式')}</span>${hazards.length?`<span class="context-chip debuff">⚠ 当前危险：${escapeHtml(hazards.map(h=>h.label).join(' / '))}</span>`:''}${ps?`<span class="context-chip debuff">自身状态：${escapeHtml(ps)}</span>`:''}`; return;
  }
  if(run.activeEvent){box.innerHTML=`<span class="context-chip">◈ ${escapeHtml(run.activeEvent.title)}</span><span class="context-chip">选择会永久写入本次世界状态</span>`;return;}
  const els=(room.elements||[]).map(el=>({el,st:elementState(run,room,el)})).filter(x=>x.st.visible);
  box.innerHTML=els.length?els.map(({el,st})=>`<button class="context-chip ${st.done?'done':''} ${st.locked&&!st.done?'locked':''}" ${st.done||st.locked?'disabled':`data-context-interact="${el.id}"`}>${st.done?'✓':st.locked?'🔒':el.action==='dialogue'?'♟':'◌'} ${escapeHtml(el.label)}${st.locked&&!st.done?` · ${escapeHtml(st.reason)}`:''}</button>`).join(''):'<span class="context-chip">这里没有明显可调查目标</span>';
  $$('[data-context-interact]').forEach(b=>b.onclick=()=>interact(b.dataset.contextInteract));
}
function adjacentRooms(run){
  const ids=[]; for(const e of run.edges){if(e.from===run.currentRoomId)ids.push(e.to);else if(e.to===run.currentRoomId)ids.push(e.from)}
  return [...new Set(ids)].map(id=>run.rooms[id]).filter(Boolean);
}
function renderActions(run,room){
  const box=$('#actionArea');
  if(run.gameOver){box.innerHTML=`<button class="action-btn primary" onclick="location.reload()">${run.victory?'完成冒险 · 返回标题':'重新开始'}</button><span class="action-hint">${run.victory?'墓城的命运已经改变。':'你的旅程停在了这里。'}</span>`;return;}
  if(run.activeDialogue){
    const d=currentDialogue(run); if(!d){box.innerHTML='<span class="action-hint">对话状态异常</span>';return;}
    let hotkey=1; const buttons=d.node.choices.map(c=>{const lock=dialogueChoiceLock(run,c);return `<button class="action-btn dialogue-choice ${c.openShop?'trade':''}" ${lock?`disabled title="${escapeHtml(lock)}"`:`data-dialogue-choice="${c.id}" data-hotkey="${hotkey++}"`}><kbd>${Math.min(9,hotkey-1)}</kbd>${c.openShop?'¤ ':''}${escapeHtml(c.text)}${lock?` · ${escapeHtml(lock)}`:''}</button>`}).join('');
    box.innerHTML=buttons+`<button class="action-btn ghost" data-dialogue-close>结束对话</button><span class="action-hint">关系 ${relationshipText(run.npcRelations?.[d.npc.id]||0)} · 你的选择会影响任务与世界事实</span>`;
    $$('[data-dialogue-choice]').forEach(b=>b.onclick=()=>dialogueChoice(b.dataset.dialogueChoice)); $('[data-dialogue-close]').onclick=closeDialogue; return;
  }
  if(run.activeShop){box.innerHTML=`<button class="action-btn primary" data-shop-tab="buy">购买</button><button class="action-btn" data-shop-tab="sell">出售</button><button class="action-btn ghost" data-shop-close>结束交易</button><span class="action-hint">古金币 ${run.player.gold} · 关键剧情物品不可出售</span>`;$$('[data-shop-tab]').forEach(b=>b.onclick=()=>{state.shopTab=b.dataset.shopTab;renderShopPanel(run)});$('[data-shop-close]').onclick=closeShop;return;}
  if(run.combat){
    const skills=(run.player.skills||[]).map(id=>state.world.skills?.[id]).filter(Boolean), potion=countItem(run,'healing_draught');
    const dist=Math.max(1,Math.min(3,run.combat.distance||2)), dn=['','近距','中距','远距'][dist]; let key=2;
    const skillButtons=skills.map(sk=>{const cd=run.combat.cooldowns?.[`skill:${sk.id}`]||0, baseCost=sk.cost||0, cost=(sk.id===(run.player.class==='warden'?'warden_smite':run.player.class==='ranger'?'ranger_pierce':'seer_burst')&&(run.player.growth?.signature_mastery||0)>=3)?Math.max(1,baseCost-1):baseCost, rangeLock=dist<(sk.minDistance||1)||(sk.maxDistance&&dist>sk.maxDistance);const hk=key++;return `<button class="action-btn skill-action" data-combat="skill:${sk.id}" data-hotkey="${hk}" ${cd>0||run.player.energy<cost||rangeLock?'disabled':''} title="${escapeHtml(sk.description||'')}"><kbd>${hk}</kbd>${sk.icon||'✦'} ${escapeHtml(sk.name)} ${cd>0?`· 冷却 ${cd}`:rangeLock?'· 距离不符':`· ${cost} 能量`}</button>`}).join('');
    const guardKey=key++,advKey=key++,retKey=key++,potKey=key++,fleeKey=key++;
    box.innerHTML=`<button class="action-btn primary" data-combat="attack" data-hotkey="1"><kbd>1</kbd>⚔ 普通攻击</button>${skillButtons}<button class="action-btn" data-combat="guard" data-hotkey="${guardKey}"><kbd>${guardKey}</kbd>⛨ 防御</button><button class="action-btn position" data-combat="advance" data-hotkey="${advKey}" ${dist<=1?'disabled':''}><kbd>${advKey}</kbd>→ 推进</button><button class="action-btn position" data-combat="retreat" data-hotkey="${retKey}" ${dist>=3?'disabled':''}><kbd>${retKey}</kbd>← 后撤</button><button class="action-btn" data-combat="potion" data-hotkey="${potKey}" ${potion<=0?'disabled':''}><kbd>${potKey}</kbd>🧪 药剂 ×${potion}</button><button class="action-btn danger" data-combat="flee" data-hotkey="${fleeKey}"><kbd>${fleeKey}</kbd>↶ 脱离战斗</button><span class="action-hint">ROUND ${run.combat.round+1} · ${dn} · ${escapeHtml(run.combat.terrainHint||'')}</span>`;
    $$('[data-combat]').forEach(b=>b.onclick=()=>combatAction(b.dataset.combat)); return;
  }
  if(run.activeEvent){
    box.innerHTML=run.activeEvent.choices.map((c,i)=>`<button class="action-btn ${c.check?'primary':''}" data-choice="${c.id}" data-hotkey="${i+1}"><kbd>${i+1}</kbd>${escapeHtml(c.text)}${c.check?` · ${attrNames[c.check.attribute]} DC${c.check.dc}`:''}</button>`).join('')+`<span class="action-hint">${escapeHtml(run.activeEvent.title)}</span>`;
    $$('[data-choice]').forEach(b=>b.onclick=()=>chooseEvent(b.dataset.choice)); return;
  }
  const els=(room.elements||[]).map(el=>({el,st:elementState(run,room,el)})).filter(x=>x.st.visible&&!x.st.done);
  const nearby=adjacentRooms(run).filter(r=>r.discovered);
  const fixedNPCs=new Set((room.elements||[]).filter(el=>el.action==='dialogue').map(el=>el.target));
  const dynamicNPCs=Object.entries(run.npcLocations||{}).filter(([id,rid])=>rid===room.id&&!fixedNPCs.has(id)).map(([id])=>state.world.npcs?.[id]).filter(Boolean);
  let hotkey=1, html='';
  for(const npc of dynamicNPCs){html+=`<button class="action-btn npc-action" data-dynamic-action-npc="${npc.id}" data-hotkey="${hotkey}"><kbd>${hotkey++}</kbd>♟ 对话 · ${escapeHtml(npc.name)} <small>${escapeHtml(npc.title)}</small></button>`;}
  for(const {el,st} of els.slice(0,5)){
    if(st.locked){html+=`<button class="action-btn locked" disabled>🔒 ${escapeHtml(el.label)} · ${escapeHtml(st.reason)}</button>`;continue;}
    html+=`<button class="action-btn ${el.action==='dialogue'?'npc-action':''}" data-action-interact="${el.id}" data-hotkey="${hotkey}"><kbd>${hotkey++}</kbd>${el.action==='dialogue'?'♟ 对话':kindNames[el.kind]||'调查'} · ${escapeHtml(el.label)}${el.check?` · DC${el.check.dc}`:''}</button>`;
  }
  for(const r of nearby){
    if(r.locked){html+=`<button class="action-btn locked" disabled>🔒 ${escapeHtml(r.visited?r.name:(typeNames[r.type]||'未知道路'))}</button>`;continue;}
    html+=`<button class="action-btn travel" data-move="${r.id}" data-hotkey="${hotkey}"><kbd>${hotkey++}</kbd>→ 前往 ${escapeHtml(r.visited?r.name:(typeNames[r.type]||'未知地点'))}</button>`;
  }
  html+=`<span class="action-hint">V0.7 · 世界事件、NPC 日程、任务分支、场景状态都由 Run 持久化驱动</span>`; box.innerHTML=html;
  $$('[data-action-interact]').forEach(b=>b.onclick=()=>interact(b.dataset.actionInteract)); $$('[data-dynamic-action-npc]').forEach(b=>b.onclick=()=>startNPCDialogue(b.dataset.dynamicActionNpc)); $$('[data-move]').forEach(b=>b.onclick=()=>moveTo(b.dataset.move));
}
async function chooseEvent(id){ sfx('interact'); await perform(`/api/runs/${state.runId}/action`,{choiceId:id},true,'event'); }
async function combatAction(action){ if(action.startsWith('skill'))actorEmote('cast');else if(action==='guard')actorEmote('guard');else if(action==='attack')actorEmote('attack');else if(action==='advance'||action==='retreat')actorEmote('idle'); sfx(action.startsWith('skill')?'skill':action==='guard'?'guard':action==='advance'||action==='retreat'?'move':'attack'); await perform(`/api/runs/${state.runId}/combat`,{action},true,'combat'); }
async function itemAction(itemId,action){ sfx(action==='equip'?'equip':'item'); await perform(`/api/runs/${state.runId}/item`,{itemId,action},false,'item'); }
async function interact(elementId){
  const run=state.snapshot.run, room=run.rooms[run.currentRoomId], el=(room.elements||[]).find(x=>x.id===elementId);
  if(el) await approachElement(el);
  sfx('interact');
  await perform(`/api/runs/${state.runId}/interact`,{elementId},true,'interact');
}
async function moveTo(roomId){ const run=state.snapshot.run;if(run.combat||run.activeEvent||run.activeDialogue||run.activeShop||run.gameOver||state.busy)return;sfx('move');await perform(`/api/runs/${state.runId}/move`,{roomId},false,'move'); }

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
    if(oldCombat&&now.combat&&oldCombat.distance!==now.combat.distance){const d=now.combat.distance||2;setActorPosition(d===1?42:d===2?28:16,74,true);spawnFloat('player',d===1?'近距':d===2?'中距':'远距','energy');}
    if(oldEnemyHP!=null){
      const nextEnemyHP=now.combat?.enemyHp ?? 0, delta=Math.max(0,oldEnemyHP-nextEnemyHP);
      if(delta>0){pulseClass('#scene','scene-hit',500);spawnFloat('enemy',`-${delta}`,now.lastRoll?.critical?'crit':'damage');if(kind==='combat'&&String(payload.action||'').startsWith('skill'))spawnSkillVfx(now.player.class);else spawnVfx(kind==='combat'?'slash':'burst','enemy');sfx(now.lastRoll?.critical?'crit':'hit');}
    }
    if(oldHP!=null&&now.player.hp<oldHP){const d=oldHP-now.player.hp;pulseClass('#scene','screen-shake',360);enemyEmote();actorEmote('hit');spawnFloat('player',`-${d}`,'damage');sfx('hurt');}
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
  $('#nearbyList').innerHTML=rooms.map(r=>`<div class="nearby-row ${r.locked?'locked':''}" ${r.locked?'':`data-nearby="${r.id}"`}><span class="nearby-icon">${r.locked?'⌧':roomIcons[r.type]||'•'}</span><div><b>${escapeHtml(r.visited?r.name:(typeNames[r.type]||'未知地点'))}</b><small>${escapeHtml(r.zone||'墓城')} · ${escapeHtml(run.regionStates?.[r.zone]||'状态未知')} · ${r.locked?'条件封锁':r.resolved?'已处理':r.visited?'仍有危险':'未踏足'}</small></div><em>${r.locked?'封锁':'前往 ›'}</em></div>`).join('')||'<p class="muted" style="font-size:9px">没有已发现道路。</p>';
  $$('[data-nearby]').forEach(n=>n.onclick=()=>moveTo(n.dataset.nearby));
}
function questStageText(stage){const m={reach_throne:'前往王座',find_captain:'寻找队长',recover_name:'寻找真名',hunt_storm_knight:'追猎风暴骑士',find_outer_gate:'寻找外封印黑门',return:'返回交付',resolved:'已产生结局',active:'进行中'};return m[stage]||stage||'';}
function questOutcomeText(outcome){const m={reported:'巡夜线重建',restored:'真名归还',route_mapped:'风暴路线已掌握',crown_broken:'灰烬王冠断裂',gate_heart_silenced:'门后心跳停止'};return m[outcome]||outcome||'';}
function renderQuests(run){
  const qs=Object.values(run.quests);$('#questList').innerHTML=qs.map(q=>{const done=q.status==='completed',stage=q.stage?` · ${questStageText(q.stage)}`:'',out=q.outcome?` · ${questOutcomeText(q.outcome)}`:'';return `<div class="quest ${done?'completed':''} ${q.status==='failed'?'failed':''}"><b>${done?'✓ ':q.status==='failed'?'× ':q.title.includes('主线')?'♛ ':'◇ '}${escapeHtml(q.title)}</b><p>${escapeHtml(q.description)}</p><small>${done?`已完成${escapeHtml(out)}`:q.status==='failed'?'已失败':`进行中 · ${q.progress} / ${q.goal}${escapeHtml(stage)}`}</small></div>`}).join('')||'<p class="muted" style="font-size:9px">暂无进行中的任务。</p>';
}
function renderWorldPulse(run){
  const c=run.clock||{bell:1,threat:0,lastChange:'第一声钟已经结束。'}, room=run.rooms[run.currentRoomId], rs=run.regionStates?.[room.zone]||'状态未知';$('#worldThreat').textContent=`威胁 ${c.threat||0} · 第 ${c.bell||1} 钟 · ${rs}`;$('#worldChange').textContent=c.lastChange||'墓城暂时保持沉默。';
}
function renderWorldEvents(run){
  const box=$('#worldEventList'); if(!box)return; const events=Object.values(run.worldEvents||{}).sort((a,b)=>(b.severity||0)-(a.severity||0));
  box.innerHTML=events.length?events.map(ev=>`<article class="world-event ${escapeHtml(ev.status||'active')} severity-${ev.severity||1}"><div><b>${ev.status==='resolved'?'✓':'◈'} ${escapeHtml(ev.title)}</b><span>${escapeHtml(ev.region)} · ${ev.stage==='escalated'?'恶化':ev.status==='resolved'?'已结束':'发展中'}</span></div><p>${escapeHtml(ev.description||'')}</p></article>`).join(''):'<p class="muted" style="font-size:9px">当前没有活跃的世界事件。</p>';
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
function openWorld(){const w=state.world;showModal(`<h2>${escapeHtml(w.title)}</h2><div class="world-info"><p>${escapeHtml(w.premise)}</p><div class="world-grid"><section class="world-box"><h3>时代</h3><p>${escapeHtml(w.era)}</p></section><section class="world-box"><h3>已知势力</h3><ul>${w.factions.map(x=>`<li>${escapeHtml(x)}</li>`).join('')}</ul></section></div><p class="muted">V0.7 Living World 在双幕战役上加入世界事件调度、NPC 日程状态、任务图分支、数据驱动技能/效果、动态场景状态、战场地形，以及可持久化内容覆盖的 GM 编辑器。</p>${state.snapshot?`<h3 style="margin-top:14px">当前区域状态</h3><div class="region-state-grid">${Object.entries(state.snapshot.run.regionStates||{}).map(([z,v])=>{const cl=/重建|回响|掌握|停止/.test(v)?'safe':/高危|暴增|破口|脉动|封闭/.test(v)?'danger':'';return `<div class="region-state-card ${cl}"><b>${escapeHtml(z)}</b><span>${escapeHtml(v)}</span></div>`}).join('')}</div>`:''}</div>`)}
function openGrowth(){
  if(!state.snapshot)return; const run=state.snapshot.run,p=run.player,cls=state.world.classes.find(c=>c.id===p.class); const talents=(state.world.talents||[]).filter(t=>t.class===p.class); const growth=state.world.growth||[];
  const attrCards=Object.entries(attrNames).map(([id,name])=>`<article class="growth-card attribute"><span>${name}</span><b>${p.attributes[id]}</b><button class="mini-btn" data-grow-attr="${id}" ${(p.attributePoints||0)<=0?'disabled':''}>+1</button></article>`).join('');
  const masteryCards=growth.map(g=>{const rank=p.growth?.[g.id]||0;return `<article class="mastery-card"><span class="talent-icon">${g.icon}</span><div><b>${escapeHtml(g.name)}</b><p>${escapeHtml(g.description)}</p><div class="rank-pips">${Array.from({length:g.maxRank},(_,i)=>`<i class="${i<rank?'on':''}"></i>`).join('')}</div></div><button class="mini-btn" data-grow-mastery="${g.id}" ${(p.masteryPoints||0)<=0||rank>=g.maxRank?'disabled':''}>训练</button></article>`}).join('');
  const talentName=Object.fromEntries(talents.map(t=>[t.id,t.name])); const talentCards=[1,2,3].map(tier=>{const rows=talents.filter(t=>(t.tier||1)===tier).map(t=>{const rank=p.talents?.[t.id]||0,learned=rank>=t.maxRank,missing=(t.requires||[]).filter(id=>(p.talents?.[id]||0)<=0),levelLock=(t.requiredLevel||0)>p.level,locked=missing.length||levelLock;const req=[];if(t.requiredLevel)req.push(`Lv.${t.requiredLevel}`);if(missing.length)req.push(`前置：${missing.map(id=>talentName[id]||id).join(' / ')}`);return `<article class="talent-card tier-${tier} ${learned?'learned':''} ${locked&&!learned?'locked':''}"><span class="talent-icon">${t.icon}</span><div><b>${escapeHtml(t.name)}</b><p>${escapeHtml(t.description)}</p><small>${learned?'已学习':req.length?escapeHtml(req.join(' · ')):`${rank} / ${t.maxRank}`}</small></div>${learned?'':`<button class="mini-btn" data-talent="${t.id}" ${(p.talentPoints||0)<=0||locked?'disabled':''}>学习</button>`}</article>`}).join('');return `<div class="talent-tier"><label>TIER ${tier}</label>${rows||'<span class="muted">暂无</span>'}</div>`}).join('');
  showModal(`<div class="growth-modal"><div class="growth-hero"><img src="${actorFramePath(p.class,'idle',1)}"><div><small>CHARACTER DEVELOPMENT</small><h2>${escapeHtml(p.name)} · ${escapeHtml(cls?.name||'职业')}</h2><p>等级 ${p.level} · 属性点 <b>${p.attributePoints||0}</b> · 精通点 <b>${p.masteryPoints||0}</b> · 天赋点 <b>${p.talentPoints||0}</b></p></div></div><h3 class="modal-section-title">基础属性</h3><div class="attribute-growth-grid">${attrCards}</div><h3 class="modal-section-title">战斗精通</h3><div class="mastery-grid">${masteryCards}</div><h3 class="modal-section-title">${escapeHtml(cls?.name||'职业')}天赋</h3><div class="talent-grid">${talentCards}</div></div>`);
  $$('[data-grow-attr]').forEach(b=>b.onclick=()=>upgradeGrowth('attribute',b.dataset.growAttr)); $$('[data-grow-mastery]').forEach(b=>b.onclick=()=>upgradeGrowth('mastery',b.dataset.growMastery)); $$('[data-talent]').forEach(b=>b.onclick=()=>learnTalent(b.dataset.talent));
}
async function learnTalent(id){
  if(state.busy)return;state.busy=true;try{state.snapshot=await api(`/api/runs/${state.runId}/talent`,{method:'POST',body:JSON.stringify({talentId:id})});render();openGrowth();toast('天赋已学习')}catch(e){toast(e.message)}finally{state.busy=false}
}
function showModal(html){$('#modalBody').innerHTML=html;$('#modal').classList.remove('hidden')}
function closeModal(){$('#modal').classList.add('hidden')}
function handleHotkeys(e){if(e.key==='Escape'&&!$('#modal').classList.contains('hidden')){closeModal();return}if(!state.snapshot||state.busy||!$('#modal').classList.contains('hidden'))return;if(/^[1-9]$/.test(e.key)){const target=document.querySelector(`[data-hotkey="${e.key}"]`);if(target&&!target.disabled){e.preventDefault();target.click()}}}


function actorFramePath(cls,mode,frame){return `/assets/actors/${cls||'warden'}_${mode||'idle'}_${frame||1}.svg`;}
function renderPlayerActor(room,run){
  const actor=$('#playerActor'), img=$('#playerActorImg'); if(!actor||!img)return;
  if(!state.actorMode)state.actorMode='idle'; img.src=actorFramePath(run.player.class,state.actorMode,state.actorFrame||1);
  if(state.actorRoom!==room.id){state.actorRoom=room.id;const d=run.combat?.distance||0;setActorPosition(run.combat?(d===1?42:d===2?28:16):16,run.combat?74:76,false);}
  else if(run.combat&&!actor.classList.contains('moving')){const d=run.combat.distance||2;setActorPosition(d===1?42:d===2?28:16,74,false);}
  actor.classList.toggle('in-combat',!!run.combat);
}
function tickActorAnimation(){
  if(!state.snapshot)return; const now=performance.now(); if(state.actorMode!=='idle'&&now>state.actorUntil){state.actorMode='idle';state.actorFrame=1;}
  state.actorFrame=(state.actorFrame%4)+1; const img=$('#playerActorImg'); if(img)img.src=actorFramePath(state.snapshot.run.player.class,state.actorMode,state.actorFrame);
}
function setActorPosition(x,y,animate=true){const a=$('#playerActor');if(!a)return;if(!animate)a.style.transition='none';a.style.left=`${Math.max(8,Math.min(88,x))}%`;a.style.top=`${Math.max(24,Math.min(82,y))}%`;if(!animate){requestAnimationFrame(()=>a.style.transition='');}}
async function approachElement(el){
  const actor=$('#playerActor');if(!actor)return;actor.classList.add('moving');setActorPosition(Math.max(12,Math.min(82,(el.x||50)-8)),Math.max(35,Math.min(78,(el.y||55)+10)),true);await new Promise(r=>setTimeout(r,380));actor.classList.remove('moving');
}
function actorEmote(kind){const a=$('#playerActor');if(!a)return;a.classList.remove('cast','hit','guard','attack');void a.offsetWidth;const mode=kind==='cast'?'cast':kind==='attack'?'attack':'idle';state.actorMode=mode;state.actorFrame=1;state.actorUntil=performance.now()+(kind==='cast'?900:kind==='attack'?720:450);a.classList.add(kind);setTimeout(()=>a.classList.remove(kind),700);}
function sceneParallax(ev){const scene=$('#scene');if(!scene)return;const r=scene.getBoundingClientRect(),x=(ev.clientX-r.left)/r.width-.5;scene.style.setProperty('--parallax-x',`${x*12}px`);scene.style.setProperty('--parallax-mid',`${x*-5}px`);scene.style.setProperty('--parallax-front',`${x*8}px`);}

function spawnFloat(target,text,kind='damage'){
  const layer=$('#vfxLayer');if(!layer)return;const el=document.createElement('div');el.className=`float-number ${kind}`;el.textContent=text;
  if(target==='enemy'){el.style.left='77%';el.style.top='43%';}else{const a=$('#playerActor');el.style.left=a?.style.left||'18%';el.style.top=a?.style.top||'73%';}
  layer.appendChild(el);setTimeout(()=>el.remove(),1150);
}
function spawnVfx(type,target='enemy'){
  const layer=$('#vfxLayer');if(!layer)return;const el=document.createElement('div');el.className=type==='burst'?'vfx-burst':type==='guard'?'vfx-guard':'vfx-slash';
  el.style.left=target==='enemy'?'68%':'16%';el.style.top=target==='enemy'?'38%':'58%';layer.appendChild(el);setTimeout(()=>el.remove(),750);
}
function spawnSkillVfx(cls){const layer=$('#vfxLayer');if(!layer)return;const el=document.createElement('div');el.className=`vfx-${cls||'warden'}`;el.style.left=cls==='ranger'?'56%':'66%';el.style.top=cls==='seer'?'30%':'42%';layer.appendChild(el);setTimeout(()=>el.remove(),900);pulseClass('#scene','skill-flare',520);}
function enemyEmote(){const p=$('#enemyPortrait');if(!p)return;p.classList.add('enemy-act');setTimeout(()=>p.classList.remove('enemy-act'),520);}
async function showCinematic(kicker,title,text,kind='',duration=1000){
  const o=$('#cinematicOverlay');if(!o)return;o.className=`cinematic-overlay show ${kind}`;$('#cinematicKicker').textContent=kicker;$('#cinematicTitle').textContent=title;$('#cinematicText').textContent=text||'';o.classList.remove('hidden');await new Promise(r=>setTimeout(r,duration));o.classList.add('hidden');o.classList.remove('show','boss','final');
}
function renderEnvironment(room,run){
  const waste=Number(room.id.split('_')[1]||0)>=33; const threat=run.clock?.threat||0, regionState=run.regionStates?.[room.zone]||'', sceneState=run.sceneStates?.[room.id]||{};
  const rear=$('#envRear'),mid=$('#envMid'),front=$('#envFront'); if(!rear||!mid||!front)return;
  rear.innerHTML=`<i class="env-orb one"></i><i class="env-orb two"></i>`;
  mid.innerHTML=Array.from({length:waste?7:5},(_,i)=>`<i class="env-particle p${i%4}"></i>`).join('');
  front.innerHTML=`<i class="env-sweep ${waste?'wind':'ash'}"></i>${threat>=2?'<i class="env-sweep omen"></i>':''}${/安全|建立|掌握/.test(regionState)||sceneState.kind==='secured'?'<i class="env-sweep safe"></i>':''}${/暴增|高危|破口/.test(regionState)||['danger','plague','blackfire','storm','echo'].includes(sceneState.kind)?`<i class="env-sweep danger ${escapeHtml(sceneState.kind||'')}"></i>`:''}`;
}
function renderSceneNPCs(room,run){
  const stage=$('#npcStage'); if(!stage)return; const can=!run.combat&&!run.activeEvent&&!run.activeDialogue&&!run.activeShop&&!run.gameOver;
  const ids=Object.entries(run.npcLocations||{}).filter(([,rid])=>rid===room.id).map(([id])=>id);
  const fixed=(room.elements||[]).filter(el=>el.action==='dialogue'&&state.world.npcs?.[el.target]); const byId=Object.fromEntries(fixed.map(el=>[el.target,el]));
  stage.innerHTML=ids.map((id,i)=>{const npc=state.world.npcs?.[id];if(!npc)return '';const base=byId[id],x=base?.x??(38+i*18),y=Math.max(23,(base?.y??48)-8),rel=run.npcRelations?.[npc.id]||0;return `<button class="scene-npc ${can?'':'disabled'}" style="left:${x}%;top:${y}%;--npc-delay:${i*.12}s" ${can?`data-dynamic-npc="${id}"`:''}><img src="${npc.portrait}" alt="${escapeHtml(npc.name)}"><span><b>${escapeHtml(npc.name)}</b><small>${escapeHtml(npc.title)} · ${relationshipText(rel)}<br>${escapeHtml(run.npcWorld?.[id]?.activity||'')}</small></span></button>`}).join('');
  $$('[data-dynamic-npc]').forEach(b=>b.onclick=()=>startNPCDialogue(b.dataset.dynamicNpc));
}
async function startNPCDialogue(npcId){if(state.busy)return;sfx('interact');await perform(`/api/runs/${state.runId}/npc`,{npcId},false,'dialogue');}

function relationshipText(v){return v>=5?'信任':v>=2?'友善':v<=-2?'戒备':v<0?'疏离':'中立';}
function currentDialogue(run){
  const a=run.activeDialogue;if(!a)return null;const npc=state.world.npcs?.[a.npcId],dlg=state.world.dialogues?.[a.dialogueId],node=dlg?.nodes?.[a.nodeId];return npc&&node?{npc,dlg,node}:null;
}
function dialogueChoiceLock(run,c){if(c.requiresFlag&&!run.flags[c.requiresFlag])return '缺少前置事实';if(c.requiresItem&&!hasItem(run,c.requiresItem))return `需要 ${state.snapshot.items[c.requiresItem]?.name||c.requiresItem}`;return '';}
function renderDialoguePanel(run){
  const p=$('#dialoguePanel');if(!p)return;const d=currentDialogue(run);if(!d){p.classList.add('hidden');return;}p.classList.remove('hidden');
  $('#dialoguePortrait').src=d.npc.portrait;$('#dialogueName').textContent=d.npc.name;$('#dialogueTitle').textContent=d.npc.title;$('#dialogueFaction').textContent=d.npc.faction;$('#dialogueText').textContent=d.node.text;
  const rel=run.npcRelations?.[d.npc.id]||0;$('#dialogueRelation').textContent=`${relationshipText(rel)} ${rel>=0?'+':''}${rel}`;
}
async function dialogueChoice(choiceId){if(state.busy)return;sfx('interact');await perform(`/api/runs/${state.runId}/dialogue`,{choiceId},false,'dialogue');}
async function closeDialogue(){if(state.busy)return;await perform(`/api/runs/${state.runId}/dialogue`,{action:'close'},false,'dialogue');}
function renderShopPanel(run){
  const panel=$('#shopPanel');if(!panel)return;const shop=state.world.shops?.[run.activeShop];if(!shop){panel.classList.add('hidden');return;}panel.classList.remove('hidden');$('#shopName').textContent=shop.name;$('#shopGold').textContent=run.player.gold;const merchant=state.world.npcs?.[shop.npcId];const sp=$('#shopPortrait');if(sp){sp.src=merchant?.portrait||'/assets/portraits/npc_broker.svg';sp.alt=merchant?.name||'商人';}const sm=$('#shopMerchantLabel');if(sm)sm.textContent=merchant?`${merchant.name} · ${merchant.title}`:'行商';
  const items=state.snapshot.items, tab=state.shopTab||'buy';
  let body='';
  if(tab==='buy') body=shop.items.map(row=>{const it=items[row.itemId],stock=run.shopStock?.[`${shop.id}:${row.itemId}`]??row.stock;return `<article class="shop-item rarity-${it?.rarity||'common'}"><span class="shop-item-icon">${it?.icon||'?'}</span><div><b>${escapeHtml(it?.name||row.itemId)}</b><small>${escapeHtml(it?.description||'')}</small><em>库存 ${stock}</em></div><button data-shop-buy="${row.itemId}" ${stock<=0||run.player.gold<row.price?'disabled':''}>◈ ${row.price}</button></article>`}).join('');
  else {const counts={};(run.player.inventory||[]).forEach(id=>counts[id]=(counts[id]||0)+1);const eq=new Set(Object.values(run.player.equipment||{}));const sellable=Object.entries(counts).filter(([id])=>{const it=items[id];return it&&['weapon','armor','trinket','consumable'].includes(it.type)&&!eq.has(id)});body=sellable.map(([id,n])=>{const it=items[id],value=Math.max(1,Math.round((it.value||1)*shop.buyback));return `<article class="shop-item sell rarity-${it.rarity||'common'}"><span class="shop-item-icon">${it.icon}</span><div><b>${escapeHtml(it.name)} ×${n}</b><small>${escapeHtml(it.description)}</small><em>回收价</em></div><button data-shop-sell="${id}">+ ◈ ${value}</button></article>`}).join('')||'<p class="shop-empty">没有可出售的非装备物品。</p>';}
  const last=run.shopRefresh?.[shop.id]||run.clock?.bell||1,next=Math.min(13,last+3),refreshHint=(run.clock?.bell||1)>=13?'墓城钟声已经停止，库存不再自动刷新':`最近补货：第 ${last} 钟 · 下一轮：第 ${next} 钟`;
  $('#shopContent').innerHTML=`<div class="shop-tabs"><button class="${tab==='buy'?'active':''}" data-shop-tab-panel="buy">购买</button><button class="${tab==='sell'?'active':''}" data-shop-tab-panel="sell">出售</button><span class="shop-refresh-hint">${escapeHtml(refreshHint)}</span></div><div class="shop-grid">${body}</div>`;
  $$('[data-shop-tab-panel]').forEach(b=>b.onclick=()=>{state.shopTab=b.dataset.shopTabPanel;renderShopPanel(run)});$$('[data-shop-buy]').forEach(b=>b.onclick=()=>shopAction('buy',b.dataset.shopBuy));$$('[data-shop-sell]').forEach(b=>b.onclick=()=>shopAction('sell',b.dataset.shopSell));
}
async function shopAction(action,itemId){if(state.busy)return;sfx(action==='buy'?'item':'equip');await perform(`/api/runs/${state.runId}/shop`,{shopId:state.snapshot.run.activeShop,action,itemId},false,'shop');}
async function closeShop(){if(state.busy)return;await perform(`/api/runs/${state.runId}/shop`,{shopId:state.snapshot.run.activeShop,action:'close'},false,'shop');state.shopTab='buy';}
async function upgradeGrowth(kind,id){if(state.busy)return;state.busy=true;try{state.snapshot=await api(`/api/runs/${state.runId}/growth`,{method:'POST',body:JSON.stringify({kind,id})});render();openGrowth();sfx('equip');toast(kind==='attribute'?'属性提升':'精通提升')}catch(e){toast(e.message);sfx('error')}finally{state.busy=false}}
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
