// Behavior of the dependency report (openspec design D5, D10 and D20, and add-report-export D6 and D7): the
// theme control, the tabs with their sliding pills, the panel motion, the copy handler and the export dialog.
// Every DOM query is guarded. Nothing is stored: the theme choice lasts for the page view only, and the
// export selection is reset on every open, so the script never touches storage.
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

	// Copy. A click on a button with data-copy copies that value through copyText. The button then shows
	// the check icon (data-state="copied") and the copied text is announced in the live region for a moment.
	const timers = new WeakMap();

	const fallbackCopy = (text) => {
		const ta = document.createElement("textarea");
		ta.value = text;
		// No readonly: iOS WebKit does not select a readonly text area, so the copy fails.
		ta.style.position = "fixed";
		ta.style.top = "0";
		ta.style.opacity = "0";
		// A modal dialog makes the rest of the page inert, so the text area must sit inside the open dialog.
		(document.querySelector("dialog[open]") || document.body).appendChild(ta);
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

	// copyText resolves once the text is copied. The clipboard API may be missing or may reject the write
	// (no permission, insecure context), so both cases fall back to the text area. Without the API the
	// fallback runs at once, inside the click, which browsers need to allow the copy command.
	const copyText = (text) => {
		if (navigator.clipboard?.writeText) {
			return navigator.clipboard
				.writeText(text)
				.catch(() => fallbackCopy(text));
		}
		fallbackCopy(text);
		return Promise.resolve();
	};

	const onCopyClick = (event) => {
		const target = event.target;
		const btn =
			target instanceof Element ? target.closest("[data-copy]") : null;
		if (!btn) {
			return;
		}
		copyText(btn.dataset.copy).then(() => flash(btn));
	};

	// Export dialog. A native modal dialog lets the user pick the repos whose updates go into a YAML export.
	// Every checkbox is checked again on each open, so a selection never outlives its dialog. Without the
	// dialog API the open button stays hidden; the entry copy buttons keep working.
	const initExport = () => {
		const openBtn = document.querySelector("button[data-export-open]");
		const dialog = document.querySelector("dialog[data-export-dialog]");
		const supported =
			typeof HTMLDialogElement !== "undefined" &&
			typeof dialog?.showModal === "function";
		if (!openBtn || !dialog || !supported) {
			if (openBtn) {
				openBtn.hidden = true;
			}
			return;
		}

		const copyBtn = dialog.querySelector("button[data-export-copy]");
		const downloadBtn = dialog.querySelector("button[data-export-download]");

		const sections = () =>
			Array.from(dialog.querySelectorAll("fieldset[data-export-section]"));
		const repoBoxes = (scope) =>
			Array.from(scope.querySelectorAll("input[data-export-repo]"));

		// The section box mirrors its repos: checked when all of them are, indeterminate when only some are.
		const syncSection = (fieldset) => {
			const sectionBox = fieldset.querySelector("input[data-export-all]");
			const repos = repoBoxes(fieldset);
			const count = repos.filter((r) => r.checked).length;
			if (sectionBox) {
				sectionBox.checked = repos.length > 0 && count === repos.length;
				sectionBox.indeterminate = count > 0 && count < repos.length;
			}
		};

		// Both actions need at least one repo, so they stay disabled while nothing is selected.
		const syncActions = () => {
			const any =
				dialog.querySelector("input[data-export-repo]:checked") !== null;
			for (const btn of [copyBtn, downloadBtn]) {
				if (btn) {
					btn.disabled = !any;
				}
			}
		};

		const syncAll = () => {
			for (const fieldset of sections()) {
				syncSection(fieldset);
			}
			syncActions();
		};

		// The export is the header, then each section with a checked repo, in DOM order. Field names are
		// fixed by the markup, so no text here is user-facing.
		const buildExport = () => {
			let text = dialog.dataset.header ?? "";
			for (const fieldset of sections()) {
				const checked = repoBoxes(fieldset).filter((r) => r.checked);
				if (checked.length === 0) {
					continue;
				}
				text += `${fieldset.dataset.key}:\n`;
				for (const repo of checked) {
					text += repo.dataset.yaml ?? "";
				}
			}
			return text;
		};

		const openExport = () => {
			if (dialog.open) {
				return;
			}
			for (const box of dialog.querySelectorAll(
				"input[data-export-all], input[data-export-repo]",
			)) {
				box.checked = true;
			}
			syncAll();
			dialog.showModal();
		};

		// One delegated listener covers every checkbox. A section box first sets its repos, then the
		// section states are recomputed, which also clears the indeterminate state.
		const onChange = (event) => {
			const target = event.target;
			if (!(target instanceof HTMLInputElement)) {
				return;
			}
			if (target.matches("input[data-export-all]")) {
				const fieldset = target.closest("fieldset[data-export-section]");
				if (fieldset) {
					for (const repo of repoBoxes(fieldset)) {
						repo.checked = target.checked;
					}
				}
				target.indeterminate = false;
			}
			syncAll();
		};

		const onCopy = () => {
			copyText(buildExport()).then(() => announce(copyBtn.dataset.copiedText));
		};

		// The link lives in the open dialog: the page behind a modal is inert, and a click on an inert link
		// may not start the download.
		const onDownload = () => {
			const url = URL.createObjectURL(
				new Blob([buildExport()], { type: "text/yaml" }),
			);
			const link = document.createElement("a");
			link.download = dialog.dataset.fileName ?? "";
			link.href = url;
			dialog.appendChild(link);
			link.click();
			link.remove();
			// Revoke after this task so the click has finished reading the URL.
			setTimeout(() => URL.revokeObjectURL(url), 0);
		};

		openBtn.addEventListener("click", openExport);
		dialog.addEventListener("change", onChange);
		// The native close (the close button, Escape) leaves focus nowhere useful, so return it to the opener.
		dialog.addEventListener("close", () => openBtn.focus());
		copyBtn?.addEventListener("click", onCopy);
		downloadBtn?.addEventListener("click", onDownload);
		// The markup may render the actions enabled; this sets them from the initial checkbox state.
		syncAll();
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
	initExport();
	placePills();
	addEventListener("resize", placePills);
	if (document.fonts) {
		// The pills are placed above already; a rejected font load keeps that placement, so the error is ignored.
		document.fonts.ready.then(placePills, () => {});
	}
})();
