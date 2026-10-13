// Behavior of the dependency report (openspec design D5, D10 and D20): the theme control, the tabs with
// their sliding pills, the panel motion and the copy handler. Every DOM query is guarded. Nothing is
// stored: the theme choice lasts for the page view only, so the script never touches storage.
(() => {
	const EASE = "cubic-bezier(.32,.72,0,1)";
	const PANEL_MS = 420;
	const PANEL_SHIFT_PX = 24;
	// How long the check icon and the announced text stay after a copy.
	const SHOW_MS = 2000;
	const STATUS_ID = "copy-status";

	const root = document.documentElement;

	const reduceMotion = () =>
		typeof matchMedia === "function" &&
		matchMedia("(prefers-reduced-motion: reduce)").matches;

	// movePill places the pill of a group under the target button.
	const movePill = (group, target) => {
		const pill = group ? group.querySelector(".pill") : null;
		if (!pill || !target) {
			return;
		}
		pill.style.width = `${target.offsetWidth}px`;
		pill.style.transform = `translateX(${target.offsetLeft}px)`;
	};

	// Theme control. The choice is light, dark or system (Automatico), and it is not stored: every page view
	// starts on system. Automatico removes data-theme, so prefers-color-scheme applies.
	const themeButtons = Array.from(
		document.querySelectorAll("[data-theme-set]"),
	);
	const themeGroup =
		themeButtons.length > 0 ? themeButtons[0].parentElement : null;

	const applyTheme = (choice) => {
		if (choice === "light" || choice === "dark") {
			root.dataset.theme = choice;
		} else {
			delete root.dataset.theme;
		}
		const shown = choice === "light" || choice === "dark" ? choice : "system";
		for (const button of themeButtons) {
			button.setAttribute(
				"aria-pressed",
				String(button.dataset.themeSet === shown),
			);
		}
	};

	const chooseTheme = (button) => {
		if (button.getAttribute("aria-pressed") === "true") {
			return;
		}
		const choice = button.dataset.themeSet;
		movePill(themeGroup, button);
		if (typeof document.startViewTransition === "function" && !reduceMotion()) {
			document.startViewTransition(() => applyTheme(choice));
		} else {
			applyTheme(choice);
		}
	};

	// Tabs. Activation is manual: a click, Enter or Space selects a tab; Left and Right only move focus.
	const tabs = Array.from(document.querySelectorAll('[role="tab"]'));
	const tabGroup = tabs.length > 0 ? tabs[0].parentElement : null;
	const panels = Array.from(document.querySelectorAll("[data-panel]"));

	const showTab = (index) => {
		const next = tabs[index];
		if (!next) {
			return;
		}
		const current = tabs.findIndex(
			(t) => t.getAttribute("aria-selected") === "true",
		);
		if (index === current) {
			return;
		}
		tabs.forEach((t, i) => {
			const on = i === index;
			t.setAttribute("aria-selected", String(on));
			t.tabIndex = on ? 0 : -1;
		});
		for (const panel of panels) {
			panel.hidden = panel.dataset.panel !== next.dataset.tab;
		}
		movePill(tabGroup, next);
		const shown = panels.find((p) => !p.hidden);
		if (shown && !reduceMotion() && typeof shown.animate === "function") {
			const dx = index > current ? PANEL_SHIFT_PX : -PANEL_SHIFT_PX;
			shown.animate(
				[
					{ opacity: 0, transform: `translateX(${dx}px)`, filter: "blur(4px)" },
					{ opacity: 1, transform: "none", filter: "blur(0)" },
				],
				{ duration: PANEL_MS, easing: EASE },
			);
		}
	};

	const focusTab = (index) => {
		const target = tabs[(index + tabs.length) % tabs.length];
		if (target) {
			target.focus();
		}
	};

	const onTabKey = (event) => {
		const i = tabs.indexOf(event.target);
		if (i < 0) {
			return;
		}
		const moves = {
			ArrowRight: i + 1,
			ArrowLeft: i - 1,
			Home: 0,
			End: tabs.length - 1,
		};
		if (event.key in moves) {
			event.preventDefault();
			focusTab(moves[event.key]);
		}
	};

	// Pills. The pill of a group follows its selected button. The placement runs without transitions.
	const placePills = () => {
		root.classList.add("no-anim");
		movePill(
			tabGroup,
			tabs.find((t) => t.getAttribute("aria-selected") === "true"),
		);
		movePill(
			themeGroup,
			themeButtons.find((b) => b.getAttribute("aria-pressed") === "true"),
		);
		requestAnimationFrame(() => {
			requestAnimationFrame(() => root.classList.remove("no-anim"));
		});
	};

	// Copy. A click on a button with data-copy copies that value, through the clipboard API or, when it
	// is missing or rejects the write, through a hidden text area. The button then shows the check icon
	// (data-state="copied") and the copied text is announced in the live region for a moment.
	const timers = new WeakMap();

	const fallbackCopy = (text) => {
		const ta = document.createElement("textarea");
		ta.value = text;
		// No readonly: iOS WebKit does not select a readonly text area, so the copy fails.
		ta.style.position = "fixed";
		ta.style.top = "0";
		ta.style.opacity = "0";
		document.body.appendChild(ta);
		ta.focus();
		ta.select();
		ta.setSelectionRange(0, text.length);
		try {
			document.execCommand("copy"); // NOSONAR iOS WebKit fallback: no non-deprecated copy path exists for it.
		} catch {
			// Without a copy command the button still shows the check icon.
		}
		ta.remove();
	};

	const announce = (text) => {
		const status = document.getElementById(STATUS_ID);
		if (status) {
			status.textContent = text;
		}
	};

	const flash = (btn) => {
		btn.dataset.state = "copied";
		if (btn.dataset.copiedOnce) {
			btn.title = btn.dataset.copiedBefore;
		}
		btn.dataset.copiedOnce = "1";
		const text = btn.dataset.copiedText;
		announce(text);
		clearTimeout(timers.get(btn));
		timers.set(
			btn,
			setTimeout(() => {
				delete btn.dataset.state;
				// Clear the announcement only when it is still this button's text.
				const status = document.getElementById(STATUS_ID);
				if (status && status.textContent === text) {
					status.textContent = "";
				}
			}, SHOW_MS),
		);
	};

	const onCopyClick = (event) => {
		const target = event.target;
		const btn =
			target instanceof Element ? target.closest("[data-copy]") : null;
		if (!btn) {
			return;
		}
		const text = btn.dataset.copy;
		if (navigator.clipboard?.writeText) {
			navigator.clipboard.writeText(text).then(
				() => flash(btn),
				() => {
					fallbackCopy(text);
					flash(btn);
				},
			);
		} else {
			fallbackCopy(text);
			flash(btn);
		}
	};

	applyTheme("system");
	for (const b of themeButtons) {
		b.addEventListener("click", () => chooseTheme(b));
	}
	tabs.forEach((t, i) => {
		t.addEventListener("click", () => showTab(i));
	});
	if (tabGroup) {
		tabGroup.addEventListener("keydown", onTabKey);
	}
	document.addEventListener("click", onCopyClick);
	placePills();
	addEventListener("resize", placePills);
	if (document.fonts) {
		// The pills are placed above already; a rejected font load keeps that placement, so the error is ignored.
		document.fonts.ready.then(placePills, () => {});
	}
})();
