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
      evt_expand: "ausklappen",
      evt_collapse: "einklappen",
      evt_error_badge: "Fehler",
      evt_thinking: "thinking",
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
      evt_expand: "expand",
      evt_collapse: "collapse",
      evt_error_badge: "error",
      evt_thinking: "thinking",
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

  const exportSession = async (pathOverride) => {
    try {
      const r = await fetch("/api/export", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(pathOverride ? { path: pathOverride } : {}),
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
      list.forEach(s => {
        const item = document.createElement("div");
        let cls = "session-item";
        if (s.current) cls += " current";
        if (s.same_project) cls += " same-project";
        if (isToday(s.last_event) || isToday(s.first_event)) cls += " today";
        if (s.id === mainID) cls += " is-main";
        item.className = cls;
        item.dataset.fullId = s.id;

        const norm = (s.first_prompt || "").replace(/\s+/g, " ").trim();
        item.dataset.popupTitle = s.short_id + (s.current ? " · " + t("session_current_badge") : "");
        item.dataset.popupBody = norm || t("session_no_prompt_tooltip");
        item.dataset.popupMeta = (s.project_label || "") + (s.first_event ? " · start " + formatRelative(s.first_event) : "");
        item.dataset.popupCmd = `ccview -s ${s.short_id}`;

        const id = document.createElement("div");
        id.className = "session-id";
        const left = document.createElement("span"); left.textContent = s.short_id;
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

        sessionList.appendChild(item);
      });
      renderFavs();
    } catch (err) {
      sessionList.innerHTML = '<div class="sidepanel-empty">' + t("list_error", { err }) + '</div>';
    }
  };

  sessionList.addEventListener("click", (e) => {
    const item = e.target.closest(".session-item");
    if (!item || !item.dataset.fullId) return;
    switchSession(item.dataset.fullId);
  });

  const activateTab = (name) => {
    tabsEl.querySelectorAll("button").forEach(b => {
      b.classList.toggle("active", b.dataset.tab === name);
    });
    tabPrompts.style.display  = name === "prompts"  ? "" : "none";
    tabSessions.style.display = name === "sessions" ? "" : "none";
    localStorage.setItem("ccview.tab", name);
    if (name === "sessions") {
      loadSessions(); // refresh every activation
      sessionsLoaded = true;
    }
  };
  tabsEl.addEventListener("click", (e) => {
    const b = e.target.closest("button[data-tab]");
    if (b) activateTab(b.dataset.tab);
  });
  activateTab(localStorage.getItem("ccview.tab") || "prompts");
  restoreLastSession();

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
    s = s.replace(/`([^`\n]+)`/g, (_, c) => `<code>${c}</code>`);
    s = s.replace(/\*\*([^*\n]+)\*\*/g, (_, c) => `<strong>${c}</strong>`);
    s = s.replace(/(^|[\s(])\*([^*\n]+)\*(?=[\s.,;:)!?]|$)/g, (_, pre, c) => `${pre}<em>${c}</em>`);
    return s;
  };

  const renderMarkdown = (text) => {
    if (!text) return "";
    const parts = text.split(/(```[a-zA-Z0-9_-]*\n[\s\S]*?```)/g);
    return parts.map(part => {
      const m = part.match(/^```([a-zA-Z0-9_-]*)\n([\s\S]*?)```$/);
      if (m) {
        const lang = m[1] ? ` data-lang="${escapeHtml(m[1])}"` : "";
        return `<pre${lang}><code>${escapeHtml(m[2].replace(/\n$/, ""))}</code></pre>`;
      }
      return part.split(/\n\n+/).filter(p => p).map(p => `<p>${renderInline(p)}</p>`).join("");
    }).join("");
  };

  const truncate = (s, n) => s.length > n ? s.slice(0, n) + "…" : s;

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
        if (input.old_string !== undefined) lines.push(`<pre><span class="diff-minus">- ${escapeHtml(truncate(input.old_string, 200))}</span></pre>`);
        if (input.new_string !== undefined) lines.push(`<pre><span class="diff-plus">+ ${escapeHtml(truncate(input.new_string, 200))}</span></pre>`);
        return lines.join("<br>");
      case "Write":
        push("file", input.file_path);
        if (input.content) lines.push(`<pre>${escapeHtml(truncate(input.content, 400))}</pre>`);
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
  const blockEl = (b) => {
    const el = document.createElement("div");
    el.className = "block " + (b.kind || "unknown");
    switch (b.kind) {
      case "text":
      case "user_prompt":
        el.innerHTML = renderMarkdown(b.text || "");
        break;
      case "thinking":
        el.textContent = b.text || "";
        break;
      case "tool_use": {
        const name = document.createElement("div");
        name.className = "tool-name";
        name.textContent = b.tool_name || "tool";
        el.appendChild(name);
        const body = document.createElement("div");
        body.innerHTML = prettyToolInput(b.tool_name, b.tool_input);
        el.appendChild(body);
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
        const body = (b.text || "").replace(/\s+$/, "");
        const long = body.split("\n").length > 10 || body.length > 500;
        if (long) el.classList.add("collapsed");
        const pre = document.createElement("pre");
        pre.textContent = body;
        el.appendChild(pre);
        if (long) {
          const btn = document.createElement("button");
          btn.className = "expand";
          btn.textContent = t("evt_expand");
          btn.onclick = () => {
            const collapsed = el.classList.toggle("collapsed");
            btn.textContent = collapsed ? t("evt_expand") : t("evt_collapse");
          };
          el.appendChild(btn);
        }
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
    (ev.blocks || []).filter(hasMeaningfulText).forEach(b => el.appendChild(blockEl(b)));
    return el;
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
