(function typedIntro() {
  const el = document.getElementById('wr-typed');
  if (!el) return;
  const full = el.getAttribute('data-full') || '';
  el.textContent = '';
  let i = 0;
  const speed = 10;
  function tick() {
    if (i <= full.length) {
      el.textContent = full.slice(0, i) + '_';
      i++;
      setTimeout(tick, speed);
    } else {
      el.textContent = full;
    }
  }
  tick();
})();

(function warRoom() {
  const leads = document.querySelectorAll('.wr-lead');
  if (!leads.length) return;

  const found = window.WR_FOUND || {};

  leads.forEach((lead) => {
    const id = lead.getAttribute('data-id');
    const head = lead.querySelector('.wr-lead-head');
    const body = document.getElementById('body-' + id);
    const status = document.getElementById('status-' + id);
    const hintBtn = lead.querySelector('[data-hint-btn="' + id + '"]');
    const hintsWrap = document.getElementById('hints-' + id);
    const submitBtn = lead.querySelector('[data-submit-btn="' + id + '"]');
    const flagInput = document.getElementById('flag-' + id);
    const feedback = document.getElementById('fb-' + id);
    const debrief = document.getElementById('debrief-' + id);
    const techniqueEl = document.getElementById('technique-' + id);
    const fixEl = document.getElementById('fix-' + id);

    if (found[id]) markSolved(lead, status);

    head.addEventListener('click', () => {
      const open = body.classList.toggle('is-open');
      if (open) {
        document.querySelectorAll('.wr-lead-body.is-open').forEach((b) => {
          if (b !== body) b.classList.remove('is-open');
        });
      }
    });

    hintBtn.addEventListener('click', async () => {
      hintBtn.disabled = true;
      hintBtn.textContent = 'Reaching out...';
      try {
        const res = await fetch('/api/hint', {
          method: 'POST', headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ id }),
        });
        const data = await res.json();
        hintsWrap.innerHTML = '';
        (data.hints || []).forEach((h, idx) => {
          const div = document.createElement('div');
          div.className = 'wr-hint';
          div.textContent = 'Thread ' + (idx + 1) + ': ' + h;
          hintsWrap.appendChild(div);
        });
        if (data.done) {
          hintBtn.textContent = 'Informant has nothing more';
        } else {
          hintBtn.disabled = false;
          hintBtn.textContent = 'Pull another thread';
        }
      } catch (e) {
        hintBtn.disabled = false;
        hintBtn.textContent = 'Pull informant thread';
      }
    });

    async function doSubmit() {
      const flag = flagInput.value.trim();
      if (!flag) return;
      submitBtn.disabled = true;
      try {
        const res = await fetch('/api/submit', {
          method: 'POST', headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ id, flag }),
        });
        const data = await res.json();
        feedback.textContent = data.message || '';
        feedback.className = 'wr-feedback ' + (data.correct ? 'is-correct' : 'is-wrong');
        if (data.correct) {
          if (data.technique) {
            techniqueEl.textContent = data.technique;
            fixEl.textContent = data.fix;
            debrief.style.display = 'flex';
          }
          markSolved(lead, status);
          updateProgress(data.found, data.total);
          if (data.all_done) setTimeout(showComplete, 500);
        }
      } catch (e) {
        feedback.textContent = 'Console unreachable - try again.';
        feedback.className = 'wr-feedback is-wrong';
      } finally {
        submitBtn.disabled = false;
      }
    }

    submitBtn.addEventListener('click', doSubmit);
    flagInput.addEventListener('keydown', (e) => { if (e.key === 'Enter') doSubmit(); });
  });

  function markSolved(lead, status) {
    lead.classList.add('is-solved');
    status.textContent = 'CONFIRMED';
  }

  function updateProgress(foundCount, total) {
    const fill = document.getElementById('wr-prog-fill');
    const label = document.getElementById('wr-prog-label');
    if (fill) fill.style.width = Math.round((foundCount / total) * 100) + '%';
    if (label) label.textContent = foundCount + ' / ' + total + ' leads closed';
  }

  function showComplete() {
    const overlay = document.getElementById('wr-complete');
    if (overlay) overlay.classList.add('is-visible');
  }
  const closeBtn = document.getElementById('wr-complete-close');
  if (closeBtn) {
    closeBtn.addEventListener('click', () => {
      document.getElementById('wr-complete').classList.remove('is-visible');
    });
  }
})();
