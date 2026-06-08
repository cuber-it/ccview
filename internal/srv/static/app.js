(() => {
  // ---------- i18n ----------
  const I18N = {
    de: {
      status_connecting: "connecting…",
      status_connected:  "verbunden",
      status_disconnected: "getrennt — reconnecting…",
      menu: "Menü",
      menu_save: "Speichern",
      menu_save_as: "Speichern unter…",
      menu_about: "Über ccview",
      menu_settings: "Einstellungen",
      settings_title: "Einstellungen",
      settings_proj_hint: "Projekt-Gruppen in der Sidebar — Anzeigename, Reihenfolge, Sichtbarkeit:",
      settings_visible: "sichtbar",
      settings_no_proj: "noch keine Projekte erkannt",
      settings_paths_title: "Projekt-Pfade",
      settings_paths_hint: "Verzeichnisse, in denen Claude-Code-Sessions liegen (eines pro Zeile):",
      settings_paths_save: "Pfade übernehmen",
      settings_paths_saved: "Übernommen — lädt neu…",
      close: "Schließen",
      tab_prompts: "Prompts",
      tab_sessions: "Sessions",
      list_none_yet: "noch keine",
      list_loading: "lade…",
      list_empty: "keine",
      prompt_filter: "filtern…",
      search_placeholder: "Suchen…  (/ öffnen, Esc schließen)",
      search_hits: "Treffer",
      sidepanel_toggle: "Panel ein-/ausklappen",
      jump_live: "Zu Live springen",
      jump_live_short: "↓ Live",
      scroll_toggle: "Auto-Scroll ein-/ausschalten",
      scroll_pause: "⏸ Scroll",
      scroll_resume: "▶ Scroll",
      go_top: "Anfang (gg)",
      go_top_short: "⇈ Anfang",
      go_live: "Live / Ende (G)",
      go_live_short: "⇊ Live",
      kbd_hint: "/ suchen · j/k Prompt · gg/G Start/Ende · Esc",
      about_tagline: "Live-Viewer für Claude Code Sessions.",
      about_built_by: "Gebaut von",
      about_license: "Lizenz:",
      empty_session: "Session ist leer. Warte auf Events…",
      evt_copy: "kopieren",
      evt_copied: "kopiert",
      evt_copy_error: "Fehler",
      evt_error_badge: "Fehler",
      evt_thinking: "thinking",
      more: "mehr",
      more_marker: "<weiter>",
      less: "weniger",
      event_user: "user",
      event_tool_result: "tool-result",
      fav_add: "Als Favorit pinnen",
      fav_remove: "Favorit entfernen",
      fav_empty: "Keine Favoriten — im Sessions-Tab pinnen (☆)",
      fav_max: "max ${n} Favoriten",
      main_add: "Als Main-Session setzen (wird beim Start geladen)",
      main_remove: "Main-Session entfernen",
      session_current_badge: "offen",
      session_no_prompt: "(kein Prompt)",
      session_no_prompt_tooltip: "(kein Prompt gefunden)",
      save_as_prompt: "Dateiname oder Pfad (leer = Default):",
      save_ok: "gespeichert: ${path}",
      save_error: "Export-Fehler: ${err}",
      switch_error: "Switch-Fehler: ${err}",
      rel_now: "jetzt",
      copy_tooltip: "Inhalt in Zwischenablage kopieren",
      list_error: "Fehler: ${err}",
    },
    en: {
      status_connecting: "connecting…",
      status_connected:  "connected",
      status_disconnected: "disconnected — reconnecting…",
      menu: "Menu",
      menu_save: "Save",
      menu_save_as: "Save As…",
      menu_about: "About ccview",
      menu_settings: "Settings",
      settings_title: "Settings",
      settings_proj_hint: "Project groups in the sidebar — display name, order, visibility:",
      settings_visible: "visible",
      settings_no_proj: "no projects detected yet",
      settings_paths_title: "Project paths",
      settings_paths_hint: "Directories where Claude Code sessions live (one per line):",
      settings_paths_save: "Apply paths",
      settings_paths_saved: "Applied — reloading…",
      close: "Close",
      tab_prompts: "Prompts",
      tab_sessions: "Sessions",
      list_none_yet: "none yet",
      list_loading: "loading…",
      list_empty: "none",
      prompt_filter: "filter…",
      search_placeholder: "Search…  (/ to open, Esc to close)",
      search_hits: "hits",
      sidepanel_toggle: "Collapse / expand panel",
      jump_live: "Jump to live",
      jump_live_short: "↓ Live",
      scroll_toggle: "Toggle auto-scroll",
      scroll_pause: "⏸ Scroll",
      scroll_resume: "▶ Scroll",
      go_top: "Top (gg)",
      go_top_short: "⇈ Top",
      go_live: "Live / End (G)",
      go_live_short: "⇊ Live",
      kbd_hint: "/ search · j/k prompt · gg/G top/end · Esc",
      about_tagline: "Live viewer for Claude Code sessions.",
      about_built_by: "Built by",
      about_license: "License:",
      empty_session: "Session is empty. Waiting for events…",
      evt_copy: "copy",
      evt_copied: "copied",
      evt_copy_error: "error",
      evt_error_badge: "error",
      evt_thinking: "thinking",
      more: "more",
      more_marker: "<more>",
      less: "less",
      event_user: "user",
      event_tool_result: "tool-result",
      fav_add: "Pin as favorite",
      fav_remove: "Unpin favorite",
      fav_empty: "No favorites — pin sessions in Sessions tab (☆)",
      fav_max: "max ${n} favorites",
      main_add: "Set as main session (loads on startup)",
      main_remove: "Clear main session",
      session_current_badge: "open",
      session_no_prompt: "(no prompt)",
      session_no_prompt_tooltip: "(no prompt found)",
      save_as_prompt: "Filename or path (blank = default):",
      save_ok: "saved: ${path}",
      save_error: "Export error: ${err}",
      switch_error: "Switch error: ${err}",
      rel_now: "now",
      copy_tooltip: "Copy to clipboard",
      list_error: "Error: ${err}",
    },
  };

  let lang = localStorage.getItem("ccview.lang") || "de";
  if (!I18N[lang]) lang = "de";

  const t = (key, vars) => {
    const s = (I18N[lang] && I18N[lang][key]) || I18N.de[key] || key;
    if (!vars) return s;
    return s.replace(/\$\{(\w+)\}/g, (_, k) => vars[k] != null ? vars[k] : "");
  };

  const applyI18n = () => {
    document.documentElement.lang = lang;
    document.querySelectorAll("[data-i18n]").forEach(el => {
      el.textContent = t(el.dataset.i18n);
    });
    document.querySelectorAll("[data-i18n-placeholder]").forEach(el => {
      el.placeholder = t(el.dataset.i18nPlaceholder);
    });
    document.querySelectorAll("[data-i18n-title]").forEach(el => {
      el.title = t(el.dataset.i18nTitle);
    });
  };

  const setLang = (newLang) => {
    if (!I18N[newLang]) return;
    lang = newLang;
    localStorage.setItem("ccview.lang", lang);
    document.querySelectorAll("#langSwitch button").forEach(b => {
      b.classList.toggle("active", b.dataset.lang === lang);
    });
    applyI18n();
    // Dynamic re-renders:
    if (typeof renderFavs === "function") renderFavs();
    if (typeof loadSessions === "function" && sessionsLoaded) loadSessions();
  };

  // ---------- theme ----------
  const html = document.documentElement;
  const switchEl = document.getElementById("themeSwitch");
  const applyTheme = (name) => {
    html.setAttribute("data-theme", name);
    localStorage.setItem("ccview.theme", name);
    switchEl.querySelectorAll("button").forEach(b => {
      b.classList.toggle("active", b.dataset.theme === name);
    });
  };
  applyTheme(localStorage.getItem("ccview.theme") || "dark");
  switchEl.addEventListener("click", (e) => {
    const btn = e.target.closest("button[data-theme]");
    if (btn) applyTheme(btn.dataset.theme);
  });

  // ---------- language switcher ----------
  const langSwitchEl = document.getElementById("langSwitch");
  langSwitchEl.addEventListener("click", (e) => {
    const btn = e.target.closest("button[data-lang]");
    if (btn) setLang(btn.dataset.lang);
  });
  // initial render
  document.querySelectorAll("#langSwitch button").forEach(b => {
    b.classList.toggle("active", b.dataset.lang === lang);
  });
  applyI18n();

  // ---------- sidepanel collapse ----------
  const panelToggle = document.getElementById("panelToggle");
  const applyPanel = (collapsed) => {
    document.body.classList.toggle("sp-collapsed", collapsed);
    panelToggle.textContent = collapsed ? "»" : "«";
    localStorage.setItem("ccview.sidepanel", collapsed ? "collapsed" : "open");
  };
  applyPanel(localStorage.getItem("ccview.sidepanel") === "collapsed");
  panelToggle.addEventListener("click", () => {
    applyPanel(!document.body.classList.contains("sp-collapsed"));
  });

  // ---------- burger menu / export ----------
  const menuToggle = document.getElementById("menuToggle");
  const menuItems = document.getElementById("menuItems");
  const openMenu = (open) => {
    menuItems.classList.toggle("open", open);
    menuToggle.classList.toggle("open", open);
  };
  menuToggle.addEventListener("click", (e) => {
    e.stopPropagation();
    openMenu(!menuItems.classList.contains("open"));
  });
  document.addEventListener("click", (e) => {
    if (!menuItems.contains(e.target) && e.target !== menuToggle) openMenu(false);
  });

  const exportSession = async (pathOverride, sessionId) => {
    try {
      const payload = {};
      if (pathOverride) payload.path = pathOverride;
      if (sessionId) payload.session = sessionId;
      const r = await fetch("/api/export", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!r.ok) throw new Error(await r.text() || r.status);
      const data = await r.json();
      setStatus(t("save_ok", { path: data.path }), "ok");
    } catch (err) {
      setStatus(t("save_error", { err }), "err");
    }
  };

  // ---------- about modal ----------
  const aboutModal = document.getElementById("aboutModal");
  const verSpan = document.getElementById("verSpan");
  const openAbout = async () => {
    try {
      const r = await fetch("/api/version");
      if (r.ok) {
        const d = await r.json();
        verSpan.textContent = "v" + (d.version || "dev");
      }
    } catch { verSpan.textContent = "vdev"; }
    aboutModal.hidden = false;
  };
  const closeAbout = () => { aboutModal.hidden = true; };
  aboutModal.addEventListener("click", (e) => {
    if (e.target.dataset.modalClose) closeAbout();
  });

  // Cheatsheet opens in its own tab so ccview stays untouched.
  const openCheatsheet = () => window.open("cheatsheet.html", "ccview-cheatsheet");
  document.addEventListener("keydown", (e) => {
    if (e.key !== "?" || e.ctrlKey || e.metaKey || e.altKey) return;
    const tag = (document.activeElement && document.activeElement.tagName) || "";
    if (tag === "INPUT" || tag === "TEXTAREA" || document.activeElement.isContentEditable) return;
    e.preventDefault();
    openCheatsheet();
  });

  // ---------- settings ----------
  const settingsModal = document.getElementById("settingsModal");
  const settingsProjects = document.getElementById("settingsProjects");
  const settingsPaths = document.getElementById("settingsPaths");
  const renderSettingsProjects = () => {
    const cfg = loadProjCfg();
    const keys = new Set([...knownProjects.keys(), ...Object.keys(cfg)]);
    const rows = [...keys].map(k => ({
      key: k,
      label: (cfg[k] && cfg[k].label) || knownProjects.get(k) || k,
      hidden: !!(cfg[k] && cfg[k].hidden),
      order: cfg[k] && cfg[k].order != null ? cfg[k].order : 1e9,
    })).sort((a, b) => a.order - b.order || a.label.localeCompare(b.label));
    settingsProjects.innerHTML = "";
    if (!rows.length) { settingsProjects.innerHTML = '<div class="sidepanel-empty">' + t("settings_no_proj") + "</div>"; return; }
    const persistOrder = () => {
      const c = loadProjCfg();
      [...settingsProjects.querySelectorAll(".settings-proj-row")].forEach((r, i) => {
        const key = r.dataset.key; c[key] = c[key] || {}; c[key].order = i;
      });
      saveProjCfg(c);
      if (typeof loadSessions === "function") loadSessions();
    };
    rows.forEach((row, idx) => {
      const el = document.createElement("div");
      el.className = "settings-proj-row"; el.dataset.key = row.key;
      const up = document.createElement("button"); up.className = "settings-ord"; up.textContent = "▲"; up.disabled = idx === 0;
      const down = document.createElement("button"); down.className = "settings-ord"; down.textContent = "▼"; down.disabled = idx === rows.length - 1;
      const name = document.createElement("input"); name.type = "text"; name.className = "settings-proj-name"; name.value = row.label;
      const vis = document.createElement("label"); vis.className = "settings-proj-vis";
      const cb = document.createElement("input"); cb.type = "checkbox"; cb.checked = !row.hidden;
      vis.append(cb, document.createTextNode(" " + t("settings_visible")));
      const persistRow = () => {
        const c = loadProjCfg(); c[row.key] = c[row.key] || {};
        c[row.key].label = name.value.trim() || knownProjects.get(row.key) || row.key;
        c[row.key].hidden = !cb.checked;
        saveProjCfg(c);
        if (typeof loadSessions === "function") loadSessions();
      };
      name.addEventListener("change", persistRow);
      cb.addEventListener("change", persistRow);
      up.addEventListener("click", () => { if (el.previousElementSibling) el.parentNode.insertBefore(el, el.previousElementSibling); persistOrder(); renderSettingsProjects(); });
      down.addEventListener("click", () => { if (el.nextElementSibling) el.parentNode.insertBefore(el.nextElementSibling, el); persistOrder(); renderSettingsProjects(); });
      el.append(up, down, name, vis);
      settingsProjects.appendChild(el);
    });
  };
  const loadRootsCfg = async () => {
    try { const r = await fetch("/api/roots"); if (r.ok) { const d = await r.json(); settingsPaths.value = (d.roots || []).join("\n"); } } catch { /* ignore */ }
  };
  const saveRootsCfg = async (btn) => {
    const roots = settingsPaths.value.split("\n").map(x => x.trim()).filter(Boolean);
    try {
      const r = await fetch("/api/roots", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ roots }) });
      if (r.ok) { btn.textContent = t("settings_paths_saved"); setTimeout(() => { btn.textContent = t("settings_paths_save"); if (typeof loadSessions === "function") loadSessions(); }, 900); }
    } catch { /* ignore */ }
  };
  const openSettings = async () => { await refreshProjGroups(); renderSettingsProjects(); loadRootsCfg(); settingsModal.hidden = false; };
  const closeSettings = () => { settingsModal.hidden = true; };
  settingsModal.addEventListener("click", (e) => { if (e.target.dataset.modalClose) closeSettings(); });
  document.getElementById("settingsPathsSave").addEventListener("click", (e) => saveRootsCfg(e.target));

  // read-only SQL query box in Settings
  const queryInput = document.getElementById("queryInput");
  const queryResult = document.getElementById("queryResult");
  const renderQueryTable = (cols, rows) => {
    queryResult.innerHTML = "";
    if (!cols.length) { queryResult.textContent = "—"; return; }
    const table = document.createElement("table");
    table.className = "query-table";
    const thead = document.createElement("thead");
    const htr = document.createElement("tr");
    cols.forEach(c => { const th = document.createElement("th"); th.textContent = c; htr.appendChild(th); });
    thead.appendChild(htr); table.appendChild(thead);
    const tbody = document.createElement("tbody");
    rows.forEach(row => {
      const tr = document.createElement("tr");
      row.forEach(v => { const td = document.createElement("td"); td.textContent = v; tr.appendChild(td); });
      tbody.appendChild(tr);
    });
    table.appendChild(tbody); queryResult.appendChild(table);
    const cnt = document.createElement("div"); cnt.className = "query-count"; cnt.textContent = rows.length + " Zeilen"; queryResult.appendChild(cnt);
  };
  const runQuery = async () => {
    const sql = (queryInput.value || "").trim();
    if (!sql) return;
    queryResult.textContent = "…";
    try {
      const r = await fetch("/api/query", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ sql }) });
      if (!r.ok) { queryResult.textContent = (await r.text()).slice(0, 200); return; }
      const d = await r.json();
      renderQueryTable(d.columns || [], d.rows || []);
    } catch { queryResult.textContent = "Fehler bei der Abfrage"; }
  };
  document.getElementById("queryRun").addEventListener("click", runQuery);

  menuItems.addEventListener("click", (e) => {
    const btn = e.target.closest("button[data-action]");
    if (!btn) return;
    openMenu(false);
    if (btn.dataset.action === "save") {
      exportSession(null);
    } else if (btn.dataset.action === "save-as") {
      const p = prompt(t("save_as_prompt"), "");
      if (p === null) return;
      exportSession(p.trim() || null);
    } else if (btn.dataset.action === "settings") {
      openSettings();
    } else if (btn.dataset.action === "cheatsheet") {
      openCheatsheet();
    } else if (btn.dataset.action === "about") {
      openAbout();
    }
  });

  // ---------- favorites ----------
  const favbarEl = document.getElementById("favbar");
  const MAX_FAVS = 5;
  let lastSessionsData = [];

  const loadFavs = () => {
    try { return JSON.parse(localStorage.getItem("ccview.favs") || "[]"); }
    catch { return []; }
  };
  const saveFavs = (favs) => localStorage.setItem("ccview.favs", JSON.stringify(favs));

  const epochOf = (iso) => iso ? new Date(iso).getTime() : 0;

  const renderFavs = () => {
    const favs = loadFavs();
    favbarEl.innerHTML = "";
    if (!favs.length) {
      const empty = document.createElement("div");
      empty.className = "favbar-empty";
      empty.textContent = t("fav_empty");
      favbarEl.appendChild(empty);
      return;
    }
    favs.forEach(fav => {
      const live = lastSessionsData.find(s => s.id === fav.id);
      const chip = document.createElement("div");
      chip.className = "fav-chip";
      if (live && live.current) chip.classList.add("active");
      const liveLast = live ? epochOf(live.last_event) : 0;
      if (liveLast > (fav.lastSeen || 0) && !(live && live.current)) chip.classList.add("updated");

      const shortEl = document.createElement("span");
      shortEl.className = "fav-short";
      shortEl.textContent = fav.shortId;
      chip.appendChild(shortEl);

      const body = document.createElement("span");
      body.className = "fav-body";
      body.textContent = fav.label || "";
      chip.appendChild(body);

      const unpin = document.createElement("button");
      unpin.className = "fav-unpin";
      unpin.textContent = "×";
      unpin.title = "Entfernen";
      unpin.addEventListener("click", (e) => {
        e.stopPropagation();
        saveFavs(loadFavs().filter(f => f.id !== fav.id));
        renderFavs();
        if (sessionsLoaded) loadSessions();
      });
      chip.appendChild(unpin);

      chip.addEventListener("click", () => {
        switchSession(fav.id);
        markFavSeen(fav.id, live ? live.last_event : null);
      });
      favbarEl.appendChild(chip);
    });
  };

  const markFavSeen = (id, iso) => {
    const favs = loadFavs();
    const f = favs.find(f => f.id === id);
    if (!f) return;
    f.lastSeen = iso ? epochOf(iso) : Date.now();
    saveFavs(favs);
    renderFavs();
  };

  const toggleFav = (session) => {
    const favs = loadFavs();
    const existing = favs.findIndex(f => f.id === session.id);
    if (existing >= 0) {
      favs.splice(existing, 1);
      saveFavs(favs);
    } else {
      if (favs.length >= MAX_FAVS) {
        setStatus(t("fav_max", { n: MAX_FAVS }), "err");
        return;
      }
      const norm = (session.first_prompt || "").replace(/\s+/g, " ").trim();
      favs.push({
        id: session.id,
        shortId: session.short_id,
        projectLabel: session.project_label,
        label: norm.length > 28 ? norm.slice(0, 28) + "..." : (norm || session.project_label || session.short_id),
        lastSeen: epochOf(session.last_event) || Date.now(),
      });
      saveFavs(favs);
    }
    renderFavs();
    loadSessions();
  };

  const isFav = (id) => loadFavs().some(f => f.id === id);

  // ---------- main session (second star, exclusive) ----------
  const getMain = () => localStorage.getItem("ccview.mainSession") || "";
  const setMain = (id) => {
    if (id) localStorage.setItem("ccview.mainSession", id);
    else localStorage.removeItem("ccview.mainSession");
  };
  const toggleMain = (sessionId) => {
    setMain(getMain() === sessionId ? "" : sessionId);
    loadSessions();
  };

  const refreshSessionsForFavs = async () => {
    if (document.hidden) return;
    try {
      const r = await fetch("/api/sessions");
      if (!r.ok) return;
      lastSessionsData = await r.json();
      // auto-update lastSeen for the currently-open favorite
      const cur = lastSessionsData.find(s => s.current);
      if (cur && isFav(cur.id)) markFavSeen(cur.id, cur.last_event);
      else renderFavs();
    } catch { /* ignore */ }
  };

  renderFavs();
  refreshSessionsForFavs();
  setInterval(refreshSessionsForFavs, 15000);

  // ---------- sidepanel tabs ----------
  const tabsEl = document.getElementById("sidepanelTabs");
  const tabPrompts = document.getElementById("tabPrompts");
  const tabSessions = document.getElementById("tabSessions");
  const tabSearch = document.getElementById("tabSearch");
  const searchAllInput = document.getElementById("searchAllInput");
  const searchAllList = document.getElementById("searchAllList");
  const sessionList = document.getElementById("sessionList");
  let sessionsLoaded = false;

  const formatRelative = (iso) => {
    if (!iso || iso.startsWith("0001-01-01")) return "";
    const d = new Date(iso);
    const diff = (Date.now() - d.getTime()) / 1000;
    if (diff < 60)     return t("rel_now");
    if (diff < 3600)   return Math.floor(diff / 60) + "m";
    if (diff < 86400)  return Math.floor(diff / 3600) + "h";
    if (diff < 604800) return Math.floor(diff / 86400) + "d";
    return d.toLocaleDateString("de-DE", { day: "2-digit", month: "2-digit" });
  };

  const isToday = (iso) => {
    if (!iso || iso.startsWith("0001-01-01")) return false;
    const d = new Date(iso);
    const n = new Date();
    return d.getFullYear() === n.getFullYear()
        && d.getMonth() === n.getMonth()
        && d.getDate() === n.getDate();
  };

  const switchSession = async (fullId) => {
    try {
      const r = await fetch("/api/switch", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id: fullId }),
      });
      if (!r.ok) throw new Error(await r.text() || r.status);
      localStorage.setItem("ccview.lastSession", fullId);
      document.dispatchEvent(new CustomEvent("ccview:session", { detail: fullId }));
      setTimeout(loadSessions, 150);
    } catch (err) {
      setStatus(t("switch_error", { err }), "err");
    }
  };

  // On startup with no active session, pick a default:
  //  1. the newest session in the current project (`same_project` flag), or
  //  2. the last-opened session from localStorage if still available.
  // Startup priority: explicit main > current project's newest > last in localStorage.
  const restoreLastSession = async () => {
    try {
      const r = await fetch("/api/sessions");
      if (!r.ok) return;
      const list = await r.json();
      lastSessionsData = list;
      const cur = list.find(s => s.current);
      if (cur) {
        localStorage.setItem("ccview.lastSession", cur.id);
        return;
      }
      const main = getMain();
      if (main && list.some(s => s.id === main)) { switchSession(main); return; }
      const sameProject = list.find(s => s.same_project);
      if (sameProject) { switchSession(sameProject.id); return; }
      const last = localStorage.getItem("ccview.lastSession");
      if (last && list.some(s => s.id === last)) switchSession(last);
    } catch { /* ignore */ }
  };

  const groupCollapsed = (key) => localStorage.getItem("ccview-grp-" + key) === "1";
  const knownProjects = new Map();
  let projGroupsCache = {};
  const loadProjCfg = () => projGroupsCache;
  const saveProjCfg = (cfg) => {
    projGroupsCache = cfg;
    const groups = Object.entries(cfg).map(([key, v]) => ({
      key, label: v.label || "", order: v.order != null ? v.order : 0, hidden: !!v.hidden,
    }));
    fetch("/api/groups", { method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ groups }) });
  };
  const refreshProjGroups = async () => {
    try {
      const r = await fetch("/api/groups");
      const d = await r.json();
      const m = {};
      (d.groups || []).forEach(g => { m[g.key] = { label: g.label, order: g.order, hidden: g.hidden }; });
      projGroupsCache = m;
    } catch { /* ignore */ }
  };
  const migrateLocalToServer = async () => {
    try {
      const names = JSON.parse(localStorage.getItem("ccview-names") || "{}");
      for (const [id, nm] of Object.entries(names)) {
        if (nm) await fetch("/api/session-meta", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ session: id, name: nm }) });
      }
      if (Object.keys(names).length) localStorage.removeItem("ccview-names");
    } catch { /* ignore */ }
    try {
      const proj = JSON.parse(localStorage.getItem("ccview-projects") || "{}");
      if (Object.keys(proj).length) {
        const groups = Object.entries(proj).map(([key, v]) => ({ key, label: v.label || "", order: v.order != null ? v.order : 0, hidden: !!v.hidden }));
        await fetch("/api/groups", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ groups }) });
        localStorage.removeItem("ccview-projects");
      }
    } catch { /* ignore */ }
  };
  migrateLocalToServer().then(refreshProjGroups).then(() => { if (typeof loadSessions === "function") loadSessions(); });
  const groupHeader = (label, count, key) => {
    const h = document.createElement("div");
    h.className = "session-group-head" + (key ? " collapsible" : "");
    if (key) h.dataset.grpKey = key;
    if (key && groupCollapsed(key)) h.classList.add("collapsed");
    h.innerHTML = (key ? '<span class="grp-caret">▾</span>' : "") +
      '<span class="grp-label"></span><span class="grp-count"></span>';
    h.querySelector(".grp-label").textContent = label;
    h.querySelector(".grp-count").textContent = count;
    return h;
  };
  const renderSessionGroups = (list, buildItem) => {
    const isActive = (s) => s.current || isToday(s.last_event) || isToday(s.first_event);
    const active = list.filter(isActive);
    const rest = list.filter(s => !isActive(s));
    list.forEach(s => knownProjects.set(s.project || "—", s.project_label || s.project || "—"));
    const addGroup = (label, items, key) => {
      const head = groupHeader(label, items.length, key);
      sessionList.appendChild(head);
      const body = document.createElement("div");
      body.className = "session-group-body";
      if (key && groupCollapsed(key)) body.style.display = "none";
      items.forEach(s => body.appendChild(buildItem(s)));
      sessionList.appendChild(body);
      if (key) head.addEventListener("click", () => {
        const willOpen = body.style.display === "none";
        body.style.display = willOpen ? "" : "none";
        localStorage.setItem("ccview-grp-" + key, willOpen ? "0" : "1");
        head.classList.toggle("collapsed", !willOpen);
      });
    };
    if (active.length) addGroup(lang === "en" ? "Active" : "Aktiv", active, null);
    const cfg = loadProjCfg();
    const groups = new Map();
    rest.forEach(s => {
      const k = s.project || "—";
      const defLabel = s.project_label || s.project || "—";
      knownProjects.set(k, defLabel);
      if (!groups.has(k)) groups.set(k, { label: defLabel, items: [] });
      groups.get(k).items.push(s);
    });
    const ordered = [...groups.entries()].sort((a, b) => {
      const oa = cfg[a[0]] && cfg[a[0]].order != null ? cfg[a[0]].order : 1e9;
      const ob = cfg[b[0]] && cfg[b[0]].order != null ? cfg[b[0]].order : 1e9;
      return oa - ob;
    });
    ordered.forEach(([k, g]) => {
      if (cfg[k] && cfg[k].hidden) return;
      addGroup((cfg[k] && cfg[k].label) || g.label, g.items, k);
    });
  };

  const sessionFilter = document.getElementById("sessionFilter");
  const applySessionFilter = () => {
    const lo = (sessionFilter.value || "").trim().toLowerCase();
    sessionList.querySelectorAll(".session-item").forEach(el => {
      el.hidden = lo && !el.textContent.toLowerCase().includes(lo);
    });
    sessionList.querySelectorAll(".session-group-body").forEach(body => {
      const head = body.previousElementSibling;
      const key = head && head.dataset ? head.dataset.grpKey : null;
      if (!lo) {
        body.style.display = (key && groupCollapsed(key)) ? "none" : "";
        if (head) head.style.display = "";
      } else {
        const any = [...body.querySelectorAll(".session-item")].some(el => !el.hidden);
        body.style.display = any ? "" : "none";
        if (head) head.style.display = any ? "" : "none";
      }
    });
  };
  sessionFilter.addEventListener("input", applySessionFilter);

  const collapseAllBtn = document.getElementById("sessionCollapseAll");
  if (collapseAllBtn) collapseAllBtn.addEventListener("click", () => {
    const heads = [...sessionList.querySelectorAll(".session-group-head.collapsible")];
    const anyOpen = heads.some(h => !h.classList.contains("collapsed"));
    heads.forEach(h => {
      const body = h.nextElementSibling;
      if (!body) return;
      body.style.display = anyOpen ? "none" : "";
      h.classList.toggle("collapsed", anyOpen);
      if (h.dataset.grpKey) localStorage.setItem("ccview-grp-" + h.dataset.grpKey, anyOpen ? "1" : "0");
    });
    collapseAllBtn.textContent = anyOpen ? "⊞" : "⊟";
  });

  const hideDoneBtn = document.getElementById("sessionHideDone");
  const applyHideDone = () => {
    const hide = localStorage.getItem("ccview-hide-done") === "1";
    document.body.classList.toggle("hide-done", hide);
    if (hideDoneBtn) hideDoneBtn.classList.toggle("active", hide);
  };
  if (hideDoneBtn) hideDoneBtn.addEventListener("click", () => {
    const hide = localStorage.getItem("ccview-hide-done") === "1";
    localStorage.setItem("ccview-hide-done", hide ? "0" : "1");
    applyHideDone();
  });
  applyHideDone();

  const loadSessions = async () => {
    try {
      const res = await fetch("/api/sessions");
      if (!res.ok) throw new Error(res.status);
      const list = await res.json();
      lastSessionsData = list;
      if (!list.length) {
        sessionList.innerHTML = '<div class="sidepanel-empty">' + t("list_empty") + '</div>';
        return;
      }
      sessionList.innerHTML = "";
      const mainID = getMain();
      const buildSessionItem = (s) => {
        const item = document.createElement("div");
        let cls = "session-item";
        if (s.current) cls += " current";
        if (s.same_project) cls += " same-project";
        if (isToday(s.last_event) || isToday(s.first_event)) cls += " today";
        if (s.id === mainID) cls += " is-main";
        if (s.done) cls += " session-done";
        item.className = cls;
        item.dataset.fullId = s.id;

        const norm = (s.first_prompt || "").replace(/\s+/g, " ").trim();
        item.dataset.popupTitle = s.short_id + (s.current ? " · " + t("session_current_badge") : "");
        item.dataset.popupBody = norm || t("session_no_prompt_tooltip");
        item.dataset.popupMeta = (s.project_label || "") + (s.first_event ? " · start " + formatRelative(s.first_event) : "");
        item.dataset.popupCmd = `ccview -s ${s.short_id}`;

        const id = document.createElement("div");
        id.className = "session-id";
        const left = document.createElement("span");
        const customName = s.name;
        left.textContent = customName || s.short_id;
        if (customName) left.classList.add("session-named");
        const right = document.createElement("span");
        right.textContent = s.current ? t("session_current_badge") : formatRelative(s.first_event || s.last_event);
        if (s.current) right.classList.add("current-badge");
        id.appendChild(left); id.appendChild(right);
        item.appendChild(id);

        const prev = document.createElement("div");
        prev.className = "session-preview";
        prev.textContent = norm ? (norm.length > 30 ? norm.slice(0, 30) + "..." : norm) : t("session_no_prompt");
        item.appendChild(prev);

        const proj = document.createElement("div");
        proj.className = "session-project";
        proj.textContent = s.project_label || s.project || "";
        item.appendChild(proj);

        const main = document.createElement("button");
        const isMain = s.id === mainID;
        main.className = "main-btn" + (isMain ? " active" : "");
        main.textContent = isMain ? "★" : "☆";
        main.title = isMain ? t("main_remove") : t("main_add");
        main.addEventListener("click", (e) => {
          e.stopPropagation();
          toggleMain(s.id);
        });
        item.appendChild(main);

        const pin = document.createElement("button");
        pin.className = "pin-btn" + (isFav(s.id) ? " pinned" : "");
        pin.textContent = isFav(s.id) ? "★" : "☆";
        pin.title = isFav(s.id) ? t("fav_remove") : t("fav_add");
        pin.addEventListener("click", (e) => {
          e.stopPropagation();
          toggleFav(s);
        });
        item.appendChild(pin);

        const burger = document.createElement("button");
        burger.className = "session-burger";
        burger.textContent = "☰";
        burger.title = "Aktionen";
        burger.addEventListener("click", (e) => {
          e.stopPropagation();
          if (!ctxMenu.hidden && ctxSession && ctxSession.id === s.id) { closeCtx(); return; }
          ctxSession = s;
          updateCtxDoneLabel();
          const rb = burger.getBoundingClientRect();
          ctxMenu.style.left = Math.max(8, Math.min(rb.right - 168, window.innerWidth - 180)) + "px";
          ctxMenu.style.top = Math.min(rb.bottom + 2, window.innerHeight - 170) + "px";
          ctxMenu.hidden = false;
        });
        item.appendChild(burger);
        return item;
      };
      renderSessionGroups(list, buildSessionItem);
      renderFavs();
      applySessionFilter();
    } catch (err) {
      sessionList.innerHTML = '<div class="sidepanel-empty">' + t("list_error", { err }) + '</div>';
    }
  };

  sessionList.addEventListener("click", (e) => {
    const item = e.target.closest(".session-item");
    if (!item || !item.dataset.fullId) return;
    switchSession(item.dataset.fullId);
  });

  // ---------- session context menu (rename / copy id / pin / favorite) ----------
  const ctxMenu = document.getElementById("sessionCtx");
  let ctxSession = null;
  const ctxDoneBtn = ctxMenu.querySelector('button[data-act="done"]');
  const updateCtxDoneLabel = () => {
    if (ctxDoneBtn && ctxSession) ctxDoneBtn.textContent = ctxSession.done ? "Einblenden" : "Ausblenden";
  };
  const setSessionName = async (id, name) => {
    await fetch("/api/session-meta", { method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ session: id, name }) });
  };
  const setSessionDone = async (id, done) => {
    await fetch("/api/session-meta", { method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ session: id, done }) });
  };
  const closeCtx = () => { ctxMenu.hidden = true; ctxSession = null; };
  sessionList.addEventListener("contextmenu", (e) => {
    const item = e.target.closest(".session-item");
    if (!item || !item.dataset.fullId) return;
    e.preventDefault();
    ctxSession = (lastSessionsData || []).find(s => s.id === item.dataset.fullId) || { id: item.dataset.fullId };
    updateCtxDoneLabel();
    ctxMenu.style.left = Math.min(e.clientX, window.innerWidth - 180) + "px";
    ctxMenu.style.top = Math.min(e.clientY, window.innerHeight - 170) + "px";
    ctxMenu.hidden = false;
  });
  ctxMenu.addEventListener("click", (e) => {
    const b = e.target.closest("button[data-act]");
    if (!b || !ctxSession) return;
    const s = ctxSession;
    if (b.dataset.act === "rename") {
      const name = prompt("Name für diese Session (leer = zurücksetzen):", s.name || "");
      if (name !== null) { setSessionName(s.id, name.trim()).then(loadSessions); }
    } else if (b.dataset.act === "done") {
      setSessionDone(s.id, !s.done).then(loadSessions);
    } else if (b.dataset.act === "copy") {
      navigator.clipboard.writeText(s.id);
    } else if (b.dataset.act === "fav") {
      toggleFav(s);
    } else if (b.dataset.act === "save") {
      const p = prompt(t("save_as_prompt"), "");
      if (p === null) return;
      exportSession(p.trim() || null, s.id);
    } else if (b.dataset.act === "delete") {
      const label = s.name || s.short_id || s.id;
      if (confirm('Session „' + label + '" löschen?\nDie .jsonl wandert in ~/.claude/ccview/trash/ (umkehrbar).')) {
        fetch("/api/delete", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ session: s.id }) })
          .then(r => { if (r.ok) loadSessions(); });
      }
    }
    closeCtx();
  });
  document.addEventListener("click", (e) => { if (!ctxMenu.hidden && !ctxMenu.contains(e.target)) closeCtx(); });
  document.addEventListener("keydown", (e) => { if (e.key === "Escape" && !ctxMenu.hidden) closeCtx(); });
  // close the menu when focus moves to another session
  sessionList.addEventListener("mouseover", (e) => {
    if (ctxMenu.hidden) return;
    const item = e.target.closest(".session-item");
    if (item && (!ctxSession || item.dataset.fullId !== ctxSession.id)) closeCtx();
  });

  const activateTab = (name) => {
    tabsEl.querySelectorAll("button").forEach(b => {
      b.classList.toggle("active", b.dataset.tab === name);
    });
    tabPrompts.style.display  = name === "prompts"  ? "" : "none";
    tabSessions.style.display = name === "sessions" ? "" : "none";
    tabSearch.style.display   = name === "search"   ? "" : "none";
    localStorage.setItem("ccview.tab", name);
    if (name === "sessions") {
      loadSessions(); // refresh every activation
      sessionsLoaded = true;
    }
    if (name === "search" && searchAllInput) searchAllInput.focus();
  };
  tabsEl.addEventListener("click", (e) => {
    const b = e.target.closest("button[data-tab]");
    if (b) activateTab(b.dataset.tab);
  });
  activateTab(localStorage.getItem("ccview.tab") || "prompts");
  restoreLastSession();

  // ---------- global regex search across all sessions ----------
  const renderSearchHits = (hits, q) => {
    if (!hits.length) { searchAllList.innerHTML = '<div class="sidepanel-empty">keine Treffer für /' + q + '/</div>'; return; }
    searchAllList.innerHTML = "";
    hits.forEach(h => {
      const days = h.days || [];
      const span = days.length ? (days[0] === days[days.length - 1] ? days[0] : days[0] + "…" + days[days.length - 1]) : "";
      const item = document.createElement("div");
      item.className = "session-item search-hit";
      item.dataset.fullId = h.id;
      const head = document.createElement("div");
      head.className = "session-id";
      const left = document.createElement("span");
      left.textContent = h.name || h.short_id;
      if (h.name) left.classList.add("session-named");
      const right = document.createElement("span");
      right.className = "search-count";
      right.textContent = h.matches + "×";
      head.append(left, right);
      item.appendChild(head);
      const meta = document.createElement("div");
      meta.className = "session-project";
      meta.textContent = (h.project_label || "") + (span ? " · " + span : "");
      item.appendChild(meta);
      if (h.snippet) {
        const sn = document.createElement("div");
        sn.className = "session-preview search-snippet";
        sn.textContent = h.snippet;
        item.appendChild(sn);
      }
      item.addEventListener("click", () => switchSession(h.id));
      searchAllList.appendChild(item);
    });
  };
  const runGlobalSearch = async () => {
    const q = searchAllInput.value.trim();
    if (!q) { searchAllList.innerHTML = '<div class="sidepanel-empty">Suchbegriff eingeben…</div>'; return; }
    searchAllList.innerHTML = '<div class="sidepanel-empty">suche…</div>';
    try {
      const r = await fetch("/api/search?q=" + encodeURIComponent(q));
      if (!r.ok) { searchAllList.innerHTML = '<div class="sidepanel-empty">' + (await r.text()).slice(0, 120) + '</div>'; return; }
      renderSearchHits(await r.json(), q);
    } catch { searchAllList.innerHTML = '<div class="sidepanel-empty">Fehler bei der Suche</div>'; }
  };
  if (searchAllInput) searchAllInput.addEventListener("keydown", (e) => { if (e.key === "Enter") runGlobalSearch(); });

  // ---------- auto-refresh session list ----------
  setInterval(() => {
    if (tabSessions.style.display !== "none" && document.visibilityState === "visible" && ctxMenu.hidden) {
      loadSessions();
    }
  }, 6000);

  // ---------- DOM refs ----------
  const eventsEl   = document.getElementById("events");
  const statusEl   = document.getElementById("status");
  const promptList = document.getElementById("promptList");
  let emptyEl = null;
  let promptCount = 0;
  let promptListEmpty = true;

  const setStatus = (text, cls) => {
    statusEl.textContent = text;
    statusEl.className = "status" + (cls ? " " + cls : "");
  };
  const showEmpty = () => {
    if (emptyEl) return;
    emptyEl = document.createElement("div");
    emptyEl.className = "empty";
    emptyEl.textContent = t("empty_session");
    eventsEl.appendChild(emptyEl);
  };
  const clearEmpty = () => { if (emptyEl) { emptyEl.remove(); emptyEl = null; } };

  const resetPromptList = () => {
    promptCount = 0;
    promptListEmpty = true;
    promptList.innerHTML = '<div class="sidepanel-empty">' + t("list_none_yet") + '</div>';
  };

  const addPromptLink = (num, text) => {
    if (promptListEmpty) { promptList.innerHTML = ""; promptListEmpty = false; }
    const a = document.createElement("a");
    a.className = "sidepanel-item";
    a.href = `#prompt-${num}`;
    const numSpan = document.createElement("span");
    numSpan.className = "num";
    numSpan.textContent = `#${String(num).padStart(4, "0")}`;
    a.appendChild(numSpan);
    const norm = (text || "").replace(/\s+/g, " ").trim();
    const preview = norm.length > 20 ? norm.slice(0, 20) + "..." : (norm || "(empty)");
    a.appendChild(document.createTextNode(preview));
    a.dataset.popupTitle = `Prompt #${String(num).padStart(4, "0")}`;
    a.dataset.popupBody = text || "(empty)";
    promptList.appendChild(a);
  };

  // ---------- markdown ----------
  const escapeHtml = (s) => s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");

  const renderInline = (text) => {
    let s = escapeHtml(text);
    // links: [text](url)
    s = s.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g,
      (_, txt, url) => `<a href="${url}" target="_blank" rel="noopener">${txt}</a>`);
    s = s.replace(/`([^`\n]+)`/g, (_, c) => `<code>${c}</code>`);
    s = s.replace(/\*\*([^*\n]+)\*\*/g, (_, c) => `<strong>${c}</strong>`);
    s = s.replace(/(^|[\s(])\*([^*\n]+)\*(?=[\s.,;:)!?]|$)/g, (_, pre, c) => `${pre}<em>${c}</em>`);
    s = s.replace(/~~([^~\n]+)~~/g, (_, c) => `<del>${c}</del>`);
    return s;
  };

  // Block-level markdown: walks lines, dispatches on heading / list / quote /
  // hr / paragraph. Fenced code blocks are sliced out by renderMarkdown first.
  const renderBlocks = (text) => {
    if (!text || !text.trim()) return "";
    const lines = text.split("\n");
    const out = [];
    const isSpecial = (l) => /^(#{1,6}\s|>|[-*+]\s|\d+\.\s|---+\s*$)/.test(l);
    let i = 0;
    while (i < lines.length) {
      const line = lines[i];
      if (!line.trim()) { i++; continue; }
      const h = line.match(/^(#{1,6})\s+(.+)$/);
      if (h) {
        out.push(`<h${h[1].length}>${renderInline(h[2])}</h${h[1].length}>`);
        i++; continue;
      }
      if (/^---+\s*$/.test(line)) { out.push("<hr>"); i++; continue; }
      if (line.startsWith(">")) {
        const buf = [];
        while (i < lines.length && lines[i].startsWith(">")) {
          buf.push(lines[i].replace(/^>\s?/, ""));
          i++;
        }
        out.push(`<blockquote>${renderBlocks(buf.join("\n"))}</blockquote>`);
        continue;
      }
      if (/^[-*+]\s/.test(line)) {
        const items = [];
        while (i < lines.length && /^[-*+]\s/.test(lines[i])) {
          items.push(`<li>${renderInline(lines[i].replace(/^[-*+]\s+/, ""))}</li>`);
          i++;
        }
        out.push(`<ul>${items.join("")}</ul>`);
        continue;
      }
      if (/^\d+\.\s/.test(line)) {
        const items = [];
        while (i < lines.length && /^\d+\.\s/.test(lines[i])) {
          items.push(`<li>${renderInline(lines[i].replace(/^\d+\.\s+/, ""))}</li>`);
          i++;
        }
        out.push(`<ol>${items.join("")}</ol>`);
        continue;
      }
      // Paragraph: collect consecutive non-empty, non-special lines.
      const buf = [];
      while (i < lines.length && lines[i].trim() && !isSpecial(lines[i])) {
        buf.push(lines[i]);
        i++;
      }
      if (buf.length) out.push(`<p>${renderInline(buf.join("\n"))}</p>`);
    }
    return out.join("");
  };

  const renderMarkdown = (text) => {
    if (!text) return "";
    // Pull out fenced code blocks first so their content isn't touched by
    // the block walker.
    const parts = text.split(/(```[a-zA-Z0-9_-]*\n[\s\S]*?```)/g);
    return parts.map(part => {
      const m = part.match(/^```([a-zA-Z0-9_-]*)\n([\s\S]*?)```$/);
      if (m) {
        const lang = m[1] ? ` data-lang="${escapeHtml(m[1])}"` : "";
        return `<pre${lang}><code>${escapeHtml(m[2].replace(/\n$/, ""))}</code></pre>`;
      }
      return renderBlocks(part);
    }).join("");
  };

  const prettyToolInput = (name, input) => {
    if (!input || typeof input !== "object") return escapeHtml(JSON.stringify(input, null, 2));
    const lines = [];
    const push = (label, value, mono = false) => {
      if (value === undefined || value === null || value === "") return;
      lines.push(`<span class="kv">${escapeHtml(label)}:</span> ${mono ? `<code>${escapeHtml(String(value))}</code>` : escapeHtml(String(value))}`);
    };
    switch (name) {
      case "Bash":  return `<pre>$ ${escapeHtml(input.command || "")}</pre>`;
      case "Read":  return `<pre>${escapeHtml(input.file_path || "")}${input.offset ? ` @${input.offset}` : ""}${input.limit ? ` limit=${input.limit}` : ""}</pre>`;
      case "Edit":
        push("file", input.file_path);
        if (input.replace_all) push("replace_all", true);
        if (input.old_string !== undefined) lines.push(`<pre><span class="diff-minus">- ${escapeHtml(input.old_string)}</span></pre>`);
        if (input.new_string !== undefined) lines.push(`<pre><span class="diff-plus">+ ${escapeHtml(input.new_string)}</span></pre>`);
        return lines.join("<br>");
      case "Write":
        push("file", input.file_path);
        if (input.content) lines.push(`<pre>${escapeHtml(input.content)}</pre>`);
        return lines.join("<br>");
      case "Grep":
        push("pattern", input.pattern, true);
        push("path", input.path); push("type", input.type); push("glob", input.glob);
        return lines.join("<br>");
      case "Glob":
        push("pattern", input.pattern, true); push("path", input.path);
        return lines.join("<br>");
      default:
        return `<pre>${escapeHtml(JSON.stringify(input, null, 2))}</pre>`;
    }
  };

  // ---------- block / event elements ----------
  // Default visual cap (in lines) per block kind. Above this, a "mehr" button
  // appears. For very long content (> STAGE2_CAP lines), a second click reveals
  // a 10-line preview marked with "<MORE>" before the third click shows all.
  const STAGE2_CAP = 10;
  const BLOCK_CAPS = {
    user_prompt: 5,
    text: 3,
    thinking: 3,
    tool_use: 3,
    tool_result: 1,
  };

  const applyClamp = (blockEl, cap) => {
    const content = blockEl.querySelector(".block-content");
    if (!content || !cap) return;
    content.style.setProperty("--clamp-lines", cap);
    content.classList.add("clamped");
    // Reading scrollHeight forces layout — we now know if content overflows.
    if (content.scrollHeight <= content.clientHeight + 2) {
      content.classList.remove("clamped");
      content.style.removeProperty("--clamp-lines");
      return;
    }
    let stage = 1; // 1 = default cap, 2 = 10-line preview, 3 = full
    const btn = document.createElement("button");
    btn.className = "more-btn";
    blockEl.appendChild(btn);
    const apply = () => {
      content.classList.remove("clamped");
      content.style.removeProperty("--clamp-lines");
      let key;
      if (stage === 1) {
        content.classList.add("clamped");
        content.style.setProperty("--clamp-lines", cap);
        key = "more";
      } else if (stage === 2) {
        content.classList.add("clamped");
        content.style.setProperty("--clamp-lines", STAGE2_CAP);
        key = "more_marker";
      } else {
        key = "less";
      }
      // data-i18n marks the button so applyI18n() picks it up on language switch
      btn.dataset.i18n = key;
      btn.textContent = t(key);
    };
    btn.onclick = () => {
      if (stage === 1) {
        // Skip stage 2 if the content already fits within the 10-line cap.
        content.classList.add("clamped");
        content.style.setProperty("--clamp-lines", STAGE2_CAP);
        stage = (content.scrollHeight > content.clientHeight + 2) ? 2 : 3;
      } else if (stage === 2) {
        stage = 3;
      } else {
        stage = 1;
      }
      apply();
    };
    apply();
  };

  const blockEl = (b) => {
    const el = document.createElement("div");
    el.className = "block " + (b.kind || "unknown");
    const makeContent = () => {
      const c = document.createElement("div");
      c.className = "block-content";
      return c;
    };
    switch (b.kind) {
      case "text":
      case "user_prompt": {
        const c = makeContent();
        c.innerHTML = renderMarkdown(b.text || "");
        el.appendChild(c);
        break;
      }
      case "thinking": {
        const c = makeContent();
        c.textContent = b.text || "";
        el.appendChild(c);
        break;
      }
      case "tool_use": {
        const name = document.createElement("div");
        name.className = "tool-name";
        name.textContent = b.tool_name || "tool";
        el.appendChild(name);
        const c = makeContent();
        c.innerHTML = prettyToolInput(b.tool_name, b.tool_input);
        el.appendChild(c);
        break;
      }
      case "image": {
        const img = document.createElement("img");
        img.alt = "(image)";
        img.style.maxWidth = "100%";
        img.style.borderRadius = "6px";
        img.style.border = "1px solid var(--border)";
        if (b.image_source === "url") {
          img.src = b.image_data || "";
        } else if (b.image_data) {
          const mt = b.image_media_type || "image/png";
          img.src = `data:${mt};base64,${b.image_data}`;
        }
        el.appendChild(img);
        break;
      }
      case "tool_result": {
        if (b.is_error) {
          el.classList.add("is-error");
          const badge = document.createElement("span");
          badge.className = "error-badge";
          badge.textContent = t("evt_error_badge");
          el.appendChild(badge);
        }
        const c = makeContent();
        const pre = document.createElement("pre");
        pre.textContent = (b.text || "").replace(/\s+$/, "");
        c.appendChild(pre);
        el.appendChild(c);
        break;
      }
      default:
        el.textContent = JSON.stringify(b);
    }
    return el;
  };

  const formatTs = (iso) => {
    if (!iso || iso.startsWith("0001-01-01")) return "";
    try {
      const d = new Date(iso);
      return d.toLocaleTimeString("de-DE", { hour: "2-digit", minute: "2-digit", second: "2-digit" });
    } catch { return ""; }
  };

  const eventToText = (ev) => {
    const parts = [];
    (ev.blocks || []).forEach(b => {
      switch (b.kind) {
        case "user_prompt":
        case "text":
          if ((b.text || "").trim()) parts.push(b.text.trim());
          break;
        case "thinking":
          if ((b.text || "").trim()) parts.push("[thinking]\n" + b.text.trim());
          break;
        case "tool_use": {
          const name = b.tool_name || "tool";
          const inp = b.tool_input;
          if (inp && typeof inp === "object") {
            if (name === "Bash" && inp.command) parts.push("[" + name + "]\n$ " + inp.command);
            else if (name === "Read" && inp.file_path) parts.push("[" + name + "] " + inp.file_path);
            else parts.push("[" + name + "]\n" + JSON.stringify(inp, null, 2));
          } else {
            parts.push("[" + name + "]");
          }
          break;
        }
        case "tool_result":
          if ((b.text || "").trim()) {
            parts.push((b.is_error ? "[result error]\n" : "[result]\n") + b.text.trim());
          }
          break;
      }
    });
    return parts.join("\n\n");
  };

  const hasMeaningfulText = (b) => {
    if (!b) return false;
    if (b.kind === "text" || b.kind === "user_prompt" || b.kind === "thinking") {
      return (b.text || "").trim().length > 0;
    }
    return true; // tool_use, tool_result, image always count
  };

  const isRealUserPrompt = (ev) =>
    ev.kind === "user" &&
    Array.isArray(ev.blocks) &&
    ev.blocks.length > 0 &&
    ev.blocks[0].kind === "user_prompt" &&
    (ev.blocks[0].text || "").trim().length > 0;

  const isToolResultUser = (ev) =>
    ev.kind === "user" &&
    Array.isArray(ev.blocks) &&
    ev.blocks.length > 0 &&
    !isRealUserPrompt(ev);

  const eventEl = (ev) => {
    const el = document.createElement("div");
    let cls = "event " + (ev.kind || "unknown");
    if (isRealUserPrompt(ev))      cls += " prompt";
    else if (isToolResultUser(ev)) cls += " tool-context";
    el.className = cls;

    const head = document.createElement("div");
    head.className = "event-header";

    const left = document.createElement("span");
    left.className = "event-label";
    if (isRealUserPrompt(ev)) {
      promptCount += 1;
      const num = promptCount;
      el.id = `prompt-${num}`;
      const badge = document.createElement("span");
      badge.className = "prompt-num";
      badge.textContent = `#${String(num).padStart(4, "0")}`;
      left.appendChild(badge);
      left.appendChild(document.createTextNode(t("event_user")));
      addPromptLink(num, ev.blocks[0].text || "");
    } else if (isToolResultUser(ev)) {
      left.textContent = t("event_tool_result");
    } else {
      left.textContent = ev.kind || "?";
    }

    const right = document.createElement("span");
    right.className = "event-right";

    const ts = document.createElement("span");
    ts.textContent = formatTs(ev.timestamp);
    right.appendChild(ts);

    const copyBtn = document.createElement("button");
    copyBtn.className = "event-copy";
    copyBtn.textContent = t("evt_copy");
    copyBtn.title = t("copy_tooltip");
    copyBtn.addEventListener("click", (e) => {
      e.stopPropagation();
      const text = eventToText(ev);
      navigator.clipboard.writeText(text).then(() => {
        copyBtn.textContent = t("evt_copied");
        copyBtn.classList.add("copied");
        setTimeout(() => {
          copyBtn.textContent = t("evt_copy");
          copyBtn.classList.remove("copied");
        }, 1200);
      }).catch(() => {
        copyBtn.textContent = t("evt_copy_error");
        setTimeout(() => { copyBtn.textContent = t("evt_copy"); }, 1500);
      });
    });
    right.appendChild(copyBtn);

    head.appendChild(left); head.appendChild(right);
    el.appendChild(head);
    (ev.blocks || []).filter(hasMeaningfulText).forEach(b => {
      el.appendChild(blockEl(b));
    });
    return el;
  };

  // Apply per-block clamps to an event AFTER it is in the document, so
  // scrollHeight/clientHeight reflect the real layout.
  const applyEventClamps = (eventNode, ev) => {
    const blocks = (ev.blocks || []).filter(hasMeaningfulText);
    const nodes = eventNode.querySelectorAll(":scope > .block");
    blocks.forEach((b, i) => {
      if (nodes[i]) applyClamp(nodes[i], BLOCK_CAPS[b.kind]);
    });
  };

  // ---------- hover popup ----------
  const popup = document.createElement("div");
  popup.className = "popup";
  popup.innerHTML = '<div class="popup-meta"><span class="popup-title"></span><span class="popup-extra"></span></div><div class="popup-body"></div><div class="popup-cmd" style="display:none"></div>';
  document.body.appendChild(popup);
  const popupTitle = popup.querySelector(".popup-title");
  const popupExtra = popup.querySelector(".popup-extra");
  const popupBody  = popup.querySelector(".popup-body");
  const popupCmd   = popup.querySelector(".popup-cmd");

  let popupTimer = null;
  const showPopup = (el) => {
    const title = el.dataset.popupTitle || "";
    const body  = el.dataset.popupBody  || "";
    const meta  = el.dataset.popupMeta  || "";
    const cmd   = el.dataset.popupCmd   || "";
    popupTitle.textContent = title;
    popupExtra.textContent = meta;
    popupBody.textContent  = body.length > 600 ? body.slice(0, 600) + " …" : body;
    if (cmd) { popupCmd.textContent = cmd; popupCmd.style.display = ""; }
    else     { popupCmd.style.display = "none"; }
    popup.classList.add("visible");
    const rect = el.getBoundingClientRect();
    const panelW = document.querySelector(".sidepanel").getBoundingClientRect().width;
    let left = panelW + 8;
    let top  = rect.top;
    // keep on screen
    if (top + 200 > window.innerHeight) top = window.innerHeight - 210;
    if (top < 50) top = 50;
    popup.style.left = left + "px";
    popup.style.top  = top  + "px";
  };
  const hidePopup = () => popup.classList.remove("visible");

  const sidepanelEl = document.querySelector(".sidepanel");
  sidepanelEl.addEventListener("mouseover", (e) => {
    const el = e.target.closest(".sidepanel-item, .session-item");
    if (!el) return;
    clearTimeout(popupTimer);
    popupTimer = setTimeout(() => showPopup(el), 180);
  });
  sidepanelEl.addEventListener("mouseout", (e) => {
    const el = e.target.closest(".sidepanel-item, .session-item");
    if (!el) return;
    clearTimeout(popupTimer);
    hidePopup();
  });

  const shouldRender = (ev) => ev.blocks && ev.blocks.some(hasMeaningfulText);

  // ---------- event counter ----------
  const statsEl = document.getElementById("stats");
  let eventCount = 0;
  const bumpStats = (delta) => {
    if (delta === 0) { eventCount = 0; }
    else eventCount += delta;
    statsEl.textContent = `${eventCount} events · ${promptCount} prompts`;
  };

  // ---------- auto-scroll-pause ----------
  const jumpLiveBtn = document.getElementById("jumpLive");
  const scrollPauseBtn = document.getElementById("scrollPauseBtn");
  const bottombar = document.getElementById("bottombar");
  let atBottom = true;
  let scrollPaused = false;
  const BOTTOM_SLOP = 60;
  const isAtBottom = () => (window.innerHeight + window.scrollY) >= document.body.scrollHeight - BOTTOM_SLOP;
  const checkBottom = () => {
    atBottom = isAtBottom();
    if (atBottom && !scrollPaused) jumpLiveBtn.classList.remove("visible");
  };
  window.addEventListener("scroll", checkBottom, { passive: true });
  const jumpToLive = () => {
    window.scrollTo(0, document.body.scrollHeight);
    atBottom = true;
    jumpLiveBtn.classList.remove("visible");
  };
  jumpLiveBtn.addEventListener("click", jumpToLive);

  const setScrollPaused = (v) => {
    scrollPaused = v;
    scrollPauseBtn.classList.toggle("active", scrollPaused);
    scrollPauseBtn.textContent = scrollPaused ? t("scroll_resume") : t("scroll_pause");
    scrollPauseBtn.title = t("scroll_toggle");
  };
  bottombar.addEventListener("click", (e) => {
    const btn = e.target.closest("button[data-action]");
    if (!btn) return;
    switch (btn.dataset.action) {
      case "scroll-pause": setScrollPaused(!scrollPaused); break;
      case "top":          window.scrollTo({ top: 0, behavior: "smooth" }); break;
      case "bottom":       jumpToLive(); break;
    }
  });

  // ---------- search ----------
  const searchbar = document.getElementById("searchbar");
  const searchInput = document.getElementById("searchInput");
  const searchCount = document.getElementById("searchCount");
  const searchClose = document.getElementById("searchClose");

  const applySearch = (q) => {
    const lo = q.trim().toLowerCase();
    const events = eventsEl.querySelectorAll(".event");
    if (!lo) {
      events.forEach(el => el.hidden = false);
      searchCount.textContent = "";
      return;
    }
    let n = 0;
    let firstMatch = null;
    events.forEach(el => {
      const txt = el.textContent.toLowerCase();
      const match = txt.includes(lo);
      el.hidden = !match;
      if (match) {
        n++;
        if (!firstMatch) firstMatch = el;
      }
    });
    searchCount.textContent = n + " " + t("search_hits");
    if (firstMatch) firstMatch.scrollIntoView({ block: "start" });
  };
  searchInput.addEventListener("input", e => applySearch(e.target.value));
  const openSearch = () => {
    searchbar.hidden = false;
    searchInput.focus();
    searchInput.select();
  };
  const closeSearch = () => {
    searchbar.hidden = true;
    if (searchInput.value) {
      searchInput.value = "";
      applySearch("");
    }
  };
  searchClose.addEventListener("click", closeSearch);

  // ---------- prompt-filter ----------
  const promptFilter = document.getElementById("promptFilter");
  promptFilter.addEventListener("input", (e) => {
    const lo = e.target.value.trim().toLowerCase();
    promptList.querySelectorAll(".sidepanel-item").forEach(el => {
      const txt = (el.textContent + " " + (el.dataset.popupBody || "")).toLowerCase();
      el.hidden = lo && !txt.includes(lo);
    });
  });

  // ---------- prompt sort (toggle ascending / descending) ----------
  const promptSortBtn = document.getElementById("promptSort");
  const applyPromptSort = () => {
    const rev = localStorage.getItem("ccview-prompt-rev") === "1";
    promptList.classList.toggle("reversed", rev);
    if (promptSortBtn) promptSortBtn.classList.toggle("active", rev);
  };
  if (promptSortBtn) {
    promptSortBtn.addEventListener("click", () => {
      const rev = localStorage.getItem("ccview-prompt-rev") === "1";
      localStorage.setItem("ccview-prompt-rev", rev ? "0" : "1");
      applyPromptSort();
    });
  }
  applyPromptSort();

  // ---------- keyboard nav ----------
  let lastKey = null;
  const jumpPrompt = (dir) => {
    const prompts = [...eventsEl.querySelectorAll(".event.prompt")].filter(el => !el.hidden);
    if (!prompts.length) return;
    const chromeH = 84;
    let idx = -1;
    for (let i = 0; i < prompts.length; i++) {
      if (prompts[i].getBoundingClientRect().top - chromeH > 4) break;
      idx = i;
    }
    const target = prompts[Math.max(0, Math.min(prompts.length - 1, idx + dir))];
    if (target) target.scrollIntoView({ behavior: "smooth", block: "start" });
  };
  document.addEventListener("keydown", (e) => {
    const t = e.target;
    const typing = t.tagName === "INPUT" || t.tagName === "TEXTAREA";
    if (e.key === "Escape") {
      if (!aboutModal.hidden) { closeAbout(); return; }
      if (!searchbar.hidden) closeSearch();
      openMenu(false);
      return;
    }
    if (typing) return;
    if (e.key === "/") { e.preventDefault(); openSearch(); return; }
    if (e.key === "j") { jumpPrompt(+1); return; }
    if (e.key === "k") { jumpPrompt(-1); return; }
    if (e.key === "G") { window.scrollTo({ top: document.body.scrollHeight, behavior: "smooth" }); return; }
    if (e.key === "g") {
      if (lastKey === "g") { window.scrollTo({ top: 0, behavior: "smooth" }); lastKey = null; }
      else { lastKey = "g"; setTimeout(() => { lastKey = null; }, 600); }
      return;
    }
    lastKey = null;
  });

  const clearView = () => {
    eventsEl.innerHTML = "";
    resetPromptList();
    emptyEl = null;
    showEmpty();
    bumpStats(0);
    atBottom = true;
    jumpLiveBtn.classList.remove("visible");
    if (promptFilter.value) promptFilter.value = "";
    if (searchInput.value) { searchInput.value = ""; searchCount.textContent = ""; }
  };

  bumpStats(0); // show "0 events · 0 prompts" on page load

  const connect = () => {
    const es = new EventSource("/stream");
    es.onopen = () => {
      setStatus(t("status_connected"), "ok");
      clearView();
    };
    es.addEventListener("reset", () => {
      clearView();
      loadSessions(); // refresh sessions tab after switch
    });
    es.onmessage = (e) => {
      clearEmpty();
      try {
        const ev = JSON.parse(e.data);
        if (!shouldRender(ev)) return;
        const node = eventEl(ev);
        // apply current search filter to new node
        if (searchInput.value) {
          const lo = searchInput.value.toLowerCase();
          node.hidden = !node.textContent.toLowerCase().includes(lo);
        }
        eventsEl.appendChild(node);
        applyEventClamps(node, ev);
        bumpStats(+1);
        if (atBottom && !scrollPaused) {
          window.scrollTo(0, document.body.scrollHeight);
        } else {
          jumpLiveBtn.classList.add("visible");
        }
      } catch (err) {
        console.error("parse error", err);
      }
    };
    es.onerror = () => setStatus(t("status_disconnected"), "err");
  };
  connect();
})();

// ---------- sidebar resize ----------
(() => {
  const handle = document.getElementById("sidebarResize");
  if (!handle) return;
  const root = document.documentElement;
  const saved = localStorage.getItem("ccview-sidebar-w");
  if (saved) root.style.setProperty("--sidebar-w", saved + "px");
  let dragging = false;
  handle.addEventListener("mousedown", (e) => {
    dragging = true; e.preventDefault(); document.body.style.userSelect = "none";
  });
  document.addEventListener("mousemove", (e) => {
    if (!dragging) return;
    root.style.setProperty("--sidebar-w", Math.max(150, Math.min(560, e.clientX)) + "px");
  });
  document.addEventListener("mouseup", () => {
    if (!dragging) return;
    dragging = false; document.body.style.userSelect = "";
    const w = parseInt(getComputedStyle(root).getPropertyValue("--sidebar-w"), 10);
    if (w) localStorage.setItem("ccview-sidebar-w", w);
  });
})();

// ---------- notes floater ----------
(() => {
  const btn = document.getElementById("notesToggle");
  const fl = document.getElementById("notesFloater");
  const ta = document.getElementById("notesText");
  const titleEl = document.getElementById("notesTitle");
  if (!btn || !fl || !ta) return;
  const easymde = new EasyMDE({
    element: ta,
    autoDownloadFontAwesome: false,
    spellChecker: false,
    status: false,
    placeholder: "Notizen zu dieser Session… (Markdown)",
    toolbar: ["bold", "italic", "heading", "|", "code", "quote", "unordered-list", "ordered-list",
      "|", "link", "table", "|", "preview", "side-by-side", "fullscreen", "|", "undo", "redo"],
  });
  let sessionId = null;
  const shortTitle = () => sessionId ? "Notizen · " + sessionId.slice(0, 8) : "Notizen";
  const load = async () => {
    sessionId = localStorage.getItem("ccview.lastSession");
    titleEl.textContent = shortTitle();
    if (!sessionId) { easymde.value(""); return; }
    try { const r = await fetch("/api/notes?session=" + encodeURIComponent(sessionId)); easymde.value((await r.json()).content || ""); } catch { /* ignore */ }
  };
  const save = async () => {
    if (!sessionId) return;
    try {
      await fetch("/api/notes?session=" + encodeURIComponent(sessionId), { method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content: easymde.value() }) });
      titleEl.textContent = "Notizen · gespeichert ✓";
      setTimeout(() => { titleEl.textContent = shortTitle(); }, 1200);
    } catch { /* ignore */ }
  };
  const open = (show) => {
    fl.hidden = !show; btn.classList.toggle("active", show);
    if (show) { load(); setTimeout(() => easymde.codemirror.refresh(), 30); }
  };
  btn.addEventListener("click", () => open(fl.hidden));
  document.getElementById("notesClose").addEventListener("click", () => open(false));
  document.getElementById("notesSave").addEventListener("click", save);
  const savedNotesW = localStorage.getItem("ccview-notes-w");
  if (savedNotesW) document.documentElement.style.setProperty("--notes-w", savedNotesW);
  const resizeEl = document.getElementById("notesResize");
  if (resizeEl) {
    let rzActive = false;
    resizeEl.addEventListener("mousedown", (e) => { rzActive = true; e.preventDefault(); document.body.style.userSelect = "none"; });
    document.addEventListener("mousemove", (e) => {
      if (!rzActive) return;
      const w = Math.max(260, Math.min(window.innerWidth - 120, window.innerWidth - e.clientX));
      document.documentElement.style.setProperty("--notes-w", w + "px");
    });
    document.addEventListener("mouseup", () => {
      if (!rzActive) return; rzActive = false; document.body.style.userSelect = "";
      localStorage.setItem("ccview-notes-w", getComputedStyle(document.documentElement).getPropertyValue("--notes-w").trim());
      easymde.codemirror.refresh();
    });
  }
  document.getElementById("notesPin").addEventListener("click", () => {
    const pinned = fl.classList.toggle("pinned");
    document.body.classList.toggle("notes-pinned", pinned);
    fl.style.left = ""; fl.style.top = ""; fl.style.right = "";
    localStorage.setItem("ccview-notes-pinned", pinned ? "1" : "0");
  });
  document.addEventListener("keydown", (e) => {
    if (!fl.hidden && (e.ctrlKey || e.metaKey) && (e.key === "s" || e.key === "S")) { e.preventDefault(); save(); }
  });
  document.addEventListener("ccview:session", () => { if (!fl.hidden) load(); });
  if (localStorage.getItem("ccview-notes-pinned") === "1") { fl.classList.add("pinned"); document.body.classList.add("notes-pinned"); }
  const head = document.getElementById("notesHead");
  let drag = null;
  head.addEventListener("mousedown", (e) => {
    if (fl.classList.contains("pinned") || e.target.closest("button")) return;
    drag = { x: e.clientX, y: e.clientY, l: fl.offsetLeft, t: fl.offsetTop }; e.preventDefault();
  });
  document.addEventListener("mousemove", (e) => {
    if (!drag) return;
    fl.style.left = (drag.l + e.clientX - drag.x) + "px";
    fl.style.top = (drag.t + e.clientY - drag.y) + "px"; fl.style.right = "auto";
  });
  document.addEventListener("mouseup", () => { drag = null; });
})();
