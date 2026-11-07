// Behavior of the dependency report (openspec design D5, D10 and D20, and add-report-export D6 and D7): the
// theme control, the tabs with their sliding pills, the panel motion, the copy handler and the export dialog.
// Optional export controls are guarded. Nothing is stored: the theme choice lasts for the page view only, and the
// export selection is reset on every open, so the script never touches storage.
(() => {
	// How long the check icon and the announced text stay after a copy.
	const SHOW_MS = 2000;
	// How long a download link's object URL stays alive. Revoking in the same task can cancel the download in
	// Safari and Firefox.
	const REVOKE_MS = 1000;

	const root = document.documentElement;

	const reduceMotion = () =>
		matchMedia("(prefers-reduced-motion: reduce)").matches;

	// movePill places the pill of a group under the target button.
	const movePill = (group, target) => {
		const pill = group.querySelector(".pill");
		const width = target.offsetWidth;
		const left = target.offsetLeft;
		pill.style.width = `${width}px`;
		pill.style.transform = `translateX(${left}px)`;
	};

	// Theme control. The choice is light, dark or system (Automatico), and it is not stored: every page view
	// starts on system. Automatico removes data-theme, so prefers-color-scheme applies.
	const themeButtons = Array.from(
		document.querySelectorAll("[data-theme-set]"),
	);
	const themeGroup = themeButtons[0].closest("[data-pill-group]");

	const applyTheme = (choice) => {
		if (choice === "light" || choice === "dark") {
			root.dataset.theme = choice;
		} else {
			delete root.dataset.theme;
		}
		const shown = choice;
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
		if (typeof document.startViewTransition === "function" && !reduceMotion()) {
			document
				.startViewTransition(() => {
					themeGroup.classList.add("no-anim");
					movePill(themeGroup, button);
					applyTheme(choice);
					// Flush the final position while transitions are disabled, before capture.
					void getComputedStyle(themeGroup.querySelector(".pill")).transform;
					themeGroup.classList.remove("no-anim");
				})
				.finished.catch(() => {});
		} else {
			movePill(themeGroup, button);
			applyTheme(choice);
		}
	};

	// Tabs. Activation is manual: a click, Enter or Space selects a tab; Left and Right only move focus.
	const tabs = Array.from(document.querySelectorAll('[role="tab"]'));
	const tabGroup = tabs[0].closest("[data-pill-group]");
	const panels = Array.from(document.querySelectorAll("[data-panel]"));

	const showTab = (index) => {
		const next = tabs[index];
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
			const show = panel.dataset.panel === next.dataset.tab;
			if (show) panel.dataset.dir = index > current ? "next" : "prev";
			panel.hidden = !show;
		}
		movePill(tabGroup, next);
	};

	const focusTab = (index) => {
		const target = tabs[(index + tabs.length) % tabs.length];
		target.focus();
	};

	const onTabKey = (event) => {
		const i = tabs.indexOf(event.target);
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

	// Reposition both pills after button sizes change, including font load and zoom.
	const pillObserver = new ResizeObserver(() => {
		const placements = [
			[tabGroup, tabs.find((t) => t.getAttribute("aria-selected") === "true")],
			[
				themeGroup,
				themeButtons.find((b) => b.getAttribute("aria-pressed") === "true"),
			],
		].map(([group, target]) => ({
			group,
			pill: group.querySelector(".pill"),
			width: target.offsetWidth,
			left: target.offsetLeft,
		}));
		for (const { group, pill, width, left } of placements) {
			group.classList.add("no-anim");
			pill.style.width = `${width}px`;
			pill.style.transform = `translateX(${left}px)`;
		}
		requestAnimationFrame(() => {
			requestAnimationFrame(() => {
				for (const { group } of placements) group.classList.remove("no-anim");
			});
		});
	});

	// Copy. A click on a button with data-copy copies that value through copyText. A successful copy flashes
	// the button: it shows the check icon or the copied text (data-state="copied") and the copied text is
	// announced in the live region for a moment.
	const timers = new WeakMap();

	// fallbackCopy returns whether the browser reports the copy as done.
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
		let copied = false;
		try {
			copied = document.execCommand("copy"); // NOSONAR iOS WebKit fallback: no non-deprecated copy path exists for it.
		} catch {
			// Without a copy command the copy failed, so the button does not flash.
		}
		ta.remove();
		return copied;
	};

	// The live region of the open export dialog when there is one: the page behind a modal is inert, so the
	// region outside the dialog is not announced then.
	const statusRegion = () =>
		document.querySelector("dialog[open] [data-export-status]") ||
		document.getElementById("copy-status");

	// The text is cleared first and set on the next frame, so a repeated text is announced again.
	const announce = (region, text) => {
		region.textContent = "";
		requestAnimationFrame(() => {
			region.textContent = text;
		});
	};

	const flash = (btn) => {
		// One region for the whole flash, so the clearing timer empties the element that was written.
		const region = statusRegion();
		btn.dataset.state = "copied";
		const copyStatus = document.getElementById("copy-status");
		if ("copy" in btn.dataset && btn.dataset.copiedOnce) {
			btn.title = copyStatus.dataset.copiedBefore;
		}
		btn.dataset.copiedOnce = "1";
		const text = btn.dataset.copiedText || copyStatus.dataset.copiedText;
		announce(region, text);
		clearTimeout(timers.get(btn));
		timers.set(
			btn,
			setTimeout(() => {
				delete btn.dataset.state;
				// Clear the announcement only when it is still this button's text.
				if (region.textContent === text) {
					region.textContent = "";
				}
			}, SHOW_MS),
		);
	};

	// copyText resolves to whether the text was copied. The clipboard API may be missing or may reject the write
	// (no permission, insecure context), so both cases fall back to the text area. Without the API the
	// fallback runs at once, inside the click, which browsers need to allow the copy command.
	const copyText = (text) => {
		if (navigator.clipboard?.writeText) {
			return navigator.clipboard.writeText(text).then(
				() => true,
				() => fallbackCopy(text),
			);
		}
		return Promise.resolve(fallbackCopy(text));
	};

	const copyAndFlash = (text, btn) => {
		void copyText(text).then((ok) => {
			if (ok) flash(btn);
		});
	};

	const onCopyClick = (event) => {
		const target = event.target;
		const btn = target.closest("[data-copy]");
		if (!btn) {
			return;
		}
		// copyText never rejects: fallbackCopy swallows errors, so the promise needs no rejection handler.
		copyAndFlash(btn.dataset.copy, btn);
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
			}
			syncAll();
		};

		const onCopy = () => {
			copyAndFlash(buildExport(), copyBtn);
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
			// Revoke later, not in this task: revoking at once can cancel the download in Safari and Firefox.
			setTimeout(() => URL.revokeObjectURL(url), REVOKE_MS);
		};

		openBtn.addEventListener("click", openExport);
		dialog.addEventListener("change", onChange);
		// The native close (the close button, Escape) leaves focus nowhere useful, so return it to the opener.
		dialog.addEventListener("close", () => openBtn.focus());
		copyBtn.addEventListener("click", onCopy);
		downloadBtn.addEventListener("click", onDownload);
		// The markup may render the actions enabled; this sets them from the initial checkbox state.
		syncAll();
	};

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
	for (const button of [...tabs, ...themeButtons]) pillObserver.observe(button);
})();
