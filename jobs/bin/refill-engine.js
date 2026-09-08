// Refill engine — the replay half of `/job apply --batch`.
//
// The batch run fills a form in a browser session that will not outlive the day. This is
// what survives it: paste the generated per-job refill.js into the console on the same
// application page (any browser, any machine, any time) and every answer goes back in.
//
// It is generated, never edited by hand: jobs/bin/make-refill.py substitutes one job's
// form-fill.json into the DATA slot below and writes
// jobs/output/applications/<job-id>/refill.js.
//
// Three rules it will not break:
//   1. It never clicks a submit button. It never clicks any button at all.
//   2. It cannot attach your resume — browsers forbid setting a file input from script.
//      It tells you which file to attach instead.
//   3. It fills only values that were written into form-fill.json by the run. It has no
//      idea what your answers are and invents nothing.

(function () {
  'use strict';

  var DATA = /*__DATA__*/ {};

  var FIELDS = DATA.fields || [];
  var out = { filled: [], missed: [], manual: [] };

  // ---- helpers -------------------------------------------------------------

  function norm(s) {
    return String(s == null ? '' : s)
      .replace(/\s+/g, ' ')
      .replace(/[*✱]/g, '')
      .replace(/\(required\)/gi, '')
      .replace(/[:：]\s*$/, '')
      .trim()
      .toLowerCase();
  }

  // Same-origin documents only. A cross-origin ATS iframe (Greenhouse embedded in a
  // careers page) is unreachable from the parent — run the script inside that frame,
  // which is what the generated header tells you to do.
  function documents() {
    var docs = [document];
    var frames = document.querySelectorAll('iframe');
    for (var i = 0; i < frames.length; i++) {
      try {
        var d = frames[i].contentDocument;
        if (d && d.body) docs.push(d);
      } catch (e) {
        /* cross-origin — skip */
      }
    }
    return docs;
  }

  function visible(el) {
    if (!el || el.disabled) return false;
    if (el.type === 'hidden') return false;
    var r = el.getBoundingClientRect();
    return !!(r.width || r.height || el.offsetParent);
  }

  // A field is found by selector first, then by its label text — ATS field ids are
  // generated and churn between page loads, but the question a human reads does not.
  function locate(f) {
    var docs = documents();
    var i, el;
    if (f.selector) {
      for (i = 0; i < docs.length; i++) {
        try {
          el = docs[i].querySelector(f.selector);
        } catch (e) {
          el = null;
        }
        if (el && visible(el)) return el;
      }
    }
    if (f.name) {
      for (i = 0; i < docs.length; i++) {
        var byName = docs[i].getElementsByName(f.name);
        for (var k = 0; k < byName.length; k++) {
          // For a radio group any member will do — fillRadio picks the right one out of
          // the group by value or by the text next to it.
          if (visible(byName[k])) return byName[k];
        }
      }
    }
    if (f.label) {
      for (i = 0; i < docs.length; i++) {
        el = byLabel(docs[i], f.label);
        if (el) return el;
      }
    }
    return null;
  }

  function byLabel(doc, label) {
    var want = norm(label);
    var labels = doc.querySelectorAll('label');
    for (var i = 0; i < labels.length; i++) {
      var txt = norm(labels[i].textContent);
      if (txt !== want && txt.indexOf(want) !== 0) continue;
      var el = null;
      var forId = labels[i].getAttribute('for');
      if (forId) el = doc.getElementById(forId);
      if (!el) el = labels[i].querySelector('input,select,textarea,[contenteditable=true]');
      if (!el) {
        // div-based forms: the control is the next field after the label
        var sib = labels[i].parentElement;
        if (sib) el = sib.querySelector('input,select,textarea,[contenteditable=true]');
      }
      if (!el && labels[i].id) {
        // combobox widgets link their label with aria-labelledby, not for=
        el = doc.querySelector('[aria-labelledby~="' + cssEscape(labels[i].id) + '"]');
      }
      if (el && visible(el)) return el;
    }
    // aria-label / placeholder fallbacks
    var all = doc.querySelectorAll('input,select,textarea,[contenteditable=true]');
    for (var j = 0; j < all.length; j++) {
      var a = norm(all[j].getAttribute('aria-label')) || norm(all[j].getAttribute('placeholder'));
      if (a && a === want && visible(all[j])) return all[j];
    }
    return null;
  }

  // React and every other controlled-input framework listen for the native setter, not
  // for el.value = x. Without this the value appears on screen and is gone on submit.
  function setValue(el, value) {
    var proto = el instanceof HTMLTextAreaElement ? HTMLTextAreaElement.prototype
      : el instanceof HTMLSelectElement ? HTMLSelectElement.prototype
        : HTMLInputElement.prototype;
    var desc = Object.getOwnPropertyDescriptor(proto, 'value');
    if (desc && desc.set) {
      desc.set.call(el, value);
    } else {
      el.value = value;
    }
    fire(el, 'input');
    fire(el, 'change');
  }

  function fire(el, type) {
    var ev;
    try {
      ev = new Event(type, { bubbles: true });
    } catch (e) {
      ev = document.createEvent('Event');
      ev.initEvent(type, true, true);
    }
    el.dispatchEvent(ev);
  }

  function fillSelect(el, value) {
    var want = norm(value);
    for (var i = 0; i < el.options.length; i++) {
      var o = el.options[i];
      if (norm(o.value) === want || norm(o.textContent) === want) {
        el.selectedIndex = i;
        fire(el, 'input');
        fire(el, 'change');
        return true;
      }
    }
    for (var j = 0; j < el.options.length; j++) {
      if (norm(el.options[j].textContent).indexOf(want) === 0) {
        el.selectedIndex = j;
        fire(el, 'input');
        fire(el, 'change');
        return true;
      }
    }
    return false;
  }

  // A checkbox is a yes/no. Its value is a state.
  function fillCheck(el, want) {
    var on = want === true || want === 'true' || norm(want) === 'yes' || norm(want) === 'on';
    return setChecked(el, on);
  }

  // A radio is a choice. Its value names WHICH button in the group to select — "No" means
  // click the No button, never "leave the group empty", which is what treating it as a
  // boolean would do.
  function fillRadio(el, want) {
    var want_ = norm(want);
    var group = el.name
      ? el.ownerDocument.querySelectorAll('input[type=radio][name="' + cssEscape(el.name) + '"]')
      : [el];
    var target = null;
    for (var i = 0; i < group.length; i++) {
      var r = group[i];
      if (norm(r.value) === want_ || norm(labelTextOf(r)) === want_) { target = r; break; }
    }
    if (!target) {
      for (var j = 0; j < group.length; j++) {
        if (norm(labelTextOf(group[j])).indexOf(want_) === 0) { target = group[j]; break; }
      }
    }
    if (!target) return false;
    return setChecked(target, true);
  }

  function setChecked(el, on) {
    if (el.checked !== on) {
      el.click(); // a real click, so framework handlers see it
      if (el.checked !== on) {
        el.checked = on;
        fire(el, 'click');
        fire(el, 'change');
      }
    }
    return el.checked === on;
  }

  // The words a human sees beside a radio or checkbox: its wrapping label, its label[for],
  // or the text right after it in div-based forms.
  function labelTextOf(el) {
    var lab = el.closest ? el.closest('label') : null;
    if (lab) return lab.textContent;
    if (el.id) {
      var f = el.ownerDocument.querySelector('label[for="' + cssEscape(el.id) + '"]');
      if (f) return f.textContent;
    }
    var sib = el.nextElementSibling;
    if (sib && !sib.querySelector('input,select,textarea')) return sib.textContent;
    return el.getAttribute('aria-label') || '';
  }

  function cssEscape(s) {
    return String(s).replace(/["\\]/g, '\\$&');
  }

  // ---- run -----------------------------------------------------------------

  var queue = [];   // comboboxes — they need the menu to open, so they run async

  FIELDS.forEach(function (f) {
    if (f.type === 'combobox') {
      queue.push(f);
      return;
    }
    if (f.type === 'file') {
      out.manual.push({ label: f.label, why: 'attach ' + (f.value || 'the file') + ' by hand — browsers block scripted file inputs' });
      return;
    }
    var el = locate(f);
    if (!el) {
      out.missed.push({ label: f.label, value: f.value, why: 'field not found on this page' });
      return;
    }
    var ok = true;
    var tag = (el.tagName || '').toLowerCase();
    if (tag === 'select') {
      ok = fillSelect(el, f.value);
    } else if (el.type === 'radio') {
      ok = fillRadio(el, f.value);
      if (!ok) out.missed.push({ label: f.label, value: f.value, why: 'no radio in that group matches' });
    } else if (el.type === 'checkbox') {
      ok = fillCheck(el, f.value);
    } else if (el.isContentEditable) {
      el.focus();
      el.textContent = String(f.value);
      fire(el, 'input');
    } else {
      setValue(el, String(f.value));
    }
    if (ok) {
      out.filled.push({ label: f.label, value: f.value, why: '' });
    } else if (el.type !== 'radio') {
      out.missed.push({ label: f.label, value: f.value, why: 'no matching option' });
    }
  });

  (DATA.gaps || []).forEach(function (g) {
    out.manual.push({ label: g.label, why: g.why || 'no answer on file — you decide' });
  });

  // A react-select combobox (Greenhouse, Ashby, Lever) is not a <select>: there is no value
  // to set. The menu has to be opened, filtered by typing, and the option clicked — and the
  // menu renders a frame later, so this half of the run is asynchronous.
  function realClick(el) {
    ['pointerdown', 'mousedown', 'mouseup', 'click'].forEach(function (t) {
      var ev;
      try {
        ev = new MouseEvent(t, { bubbles: true, cancelable: true, view: window });
      } catch (e) {
        ev = document.createEvent('MouseEvents');
        ev.initEvent(t, true, true);
      }
      el.dispatchEvent(ev);
    });
  }

  function options(doc) {
    return doc.querySelectorAll('[id^="react-select"][id*="option"], [role="option"]');
  }

  function fillCombo(f) {
    return new Promise(function (done) {
      var el = locate(f);
      if (!el) {
        out.missed.push({ label: f.label, value: f.value, why: 'combobox not found on this page' });
        return done();
      }
      var doc = el.ownerDocument;
      var control = el.closest('[class*="control"]') || el.parentElement;
      el.focus();
      realClick(control);
      setTimeout(function () {
        // typing filters the menu; some widgets need it, none mind it
        try { setValue(el, String(f.value)); } catch (e) { /* not a text input */ }
        setTimeout(function () {
          var want = norm(f.value);
          var opts = options(doc);
          var hit = null;
          for (var i = 0; i < opts.length; i++) {
            if (norm(opts[i].textContent) === want) { hit = opts[i]; break; }
          }
          if (!hit) {
            for (var j = 0; j < opts.length; j++) {
              if (norm(opts[j].textContent).indexOf(want) === 0) { hit = opts[j]; break; }
            }
          }
          if (!hit) {
            out.missed.push({ label: f.label, value: f.value, why: 'no option matches in the dropdown' });
            el.blur();
            return done();
          }
          realClick(hit);
          setTimeout(function () {
            out.filled.push({ label: f.label, value: f.value, why: '' });
            done();
          }, 90);
        }, 140);
      }, 90);
    });
  }

  // ---- report --------------------------------------------------------------

  var title = (DATA.role || 'Application') + (DATA.company ? ' @ ' + DATA.company : '');
  try {
    console.group('%c' + title, 'font-weight:700');
    console.log('filled ' + out.filled.length + ' · missed ' + out.missed.length + ' · needs you ' + out.manual.length);
    if (out.filled.length) console.table(out.filled);
    if (out.missed.length) console.table(out.missed);
    if (out.manual.length) console.table(out.manual);
    console.log('%cNothing was submitted. Review every field, attach the resume, then submit it yourself.', 'color:#b45309');
    console.groupEnd();
  } catch (e) { /* older consoles */ }

  if (!queue.length) {
    toast(title, out, DATA);
    return out;
  }

  // Serial: two menus open at once fight over focus.
  var chain = Promise.resolve();
  queue.forEach(function (f) { chain = chain.then(function () { return fillCombo(f); }); });
  chain.then(function () {
    console.log('dropdowns done — filled ' + out.filled.length + ' · missed ' + out.missed.length);
    toast(title, out, DATA);
  });
  return out;

  function toast(title, out, data) {
    try {
      var old = document.getElementById('job-refill-toast');
      if (old) old.remove();
      var box = document.createElement('div');
      box.id = 'job-refill-toast';
      box.style.cssText = 'position:fixed;z-index:2147483647;right:16px;bottom:16px;width:330px;' +
        'background:#fff;color:#18181b;border:1px solid #d4d4d8;border-radius:10px;padding:12px 14px;' +
        'box-shadow:0 12px 34px rgba(0,0,0,.22);font:13px/1.5 ui-sans-serif,-apple-system,"Segoe UI",sans-serif';
      var rows = '';
      out.manual.concat(out.missed).slice(0, 8).forEach(function (m) {
        rows += '<li style="margin:2px 0">' + esc(m.label || '(field)') +
          ' <span style="color:#71717a">— ' + esc(m.why) + '</span></li>';
      });
      box.innerHTML =
        '<div style="font-weight:700;margin-bottom:2px">' + esc(title) + '</div>' +
        '<div style="color:#3f3f46">Filled <b>' + out.filled.length + '</b> fields · ' +
        out.missed.length + ' not found · ' + out.manual.length + ' need you</div>' +
        (rows ? '<ul style="margin:8px 0 0;padding-left:18px;max-height:180px;overflow:auto">' + rows + '</ul>' : '') +
        (data.resume_file ? '<div style="margin-top:8px;color:#b45309">Attach: ' + esc(data.resume_file) + '</div>' : '') +
        '<div style="margin-top:8px;color:#b45309;font-weight:600">Not submitted — review, then submit yourself.</div>' +
        '<button id="job-refill-x" style="margin-top:9px;font:inherit;border:1px solid #d4d4d8;' +
        'background:#fafafa;border-radius:6px;padding:3px 9px;cursor:pointer">close</button>';
      document.body.appendChild(box);
      document.getElementById('job-refill-x').onclick = function () { box.remove(); };
    } catch (e) { /* a page that fights the DOM still got its console report */ }
  }

  function esc(s) {
    return String(s == null ? '' : s).replace(/[&<>"]/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[c];
    });
  }
})();
