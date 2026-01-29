(function () {
    const links = document.querySelectorAll('.site-nav .nav-link');
    const sections = Array.from(document.querySelectorAll('section[id]'));

    // Smooth scroll
    links.forEach(link => {
        link.addEventListener('click', (e) => {
            const href = link.getAttribute('href');
            if (href && href.startsWith('#')) {
                e.preventDefault();
                const target = document.querySelector(href);
                if (target) {
                    window.scrollTo({
                        top: target.offsetTop - 72,
                        behavior: 'smooth'
                    });
                }
            }
        });
    });

    // Scroll spy + reveal
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            const id = entry.target.getAttribute('id');
            if (entry.isIntersecting) {
                entry.target.classList.add('visible');
                links.forEach(l => {
                    const sec = l.dataset.section;
                    l.classList.toggle('active', sec === id);
                });
            }
        });
    }, {
        threshold: 0.35
    });

    sections.forEach(sec => observer.observe(sec));


    // Workstation launch boot sequence
    const launchBtn = document.getElementById('btn-launch-workstation');
    const bootOverlay = document.getElementById('workstation-boot-overlay');
    const bootCancel = document.getElementById('boot-cancel-btn');
    let bootInProgress = false;

    function openWorkstationPopup() {
        const w = window.screen.availWidth || 1440;
        const h = window.screen.availHeight || 900;
        const features = [
            'toolbar=no',
            'menubar=no',
            'location=no',
            'status=no',
            'scrollbars=yes',
            'resizable=yes',
            'width=' + w,
            'height=' + h,
            'top=0',
            'left=0'
        ].join(',');
        window.open('/workstation/', 'reservechain_workstation', features);
    }

    function runBootSequence() {
        if (!bootOverlay || bootInProgress) return;
        bootInProgress = true;
        bootOverlay.classList.remove('boot-overlay--hidden');
        bootOverlay.classList.add('boot-overlay--visible');

        const steps = Array.from(bootOverlay.querySelectorAll('.boot-step'));
        steps.forEach(step => {
            step.classList.remove('boot-step--active', 'boot-step--done');
            const st = step.querySelector('.boot-step__status');
            if (st) st.textContent = 'Pending';
        });

        let idx = 0;
        function advance() {
            if (!bootInProgress) return;
            if (idx > 0 && steps[idx - 1]) {
                steps[idx - 1].classList.remove('boot-step--active');
                steps[idx - 1].classList.add('boot-step--done');
                const prevStatus = steps[idx - 1].querySelector('.boot-step__status');
                if (prevStatus) prevStatus.textContent = 'OK';
            }
            if (idx < steps.length) {
                const step = steps[idx];
                step.classList.add('boot-step--active');
                const st = step.querySelector('.boot-step__status');
                if (st) {
                    if (idx === 2) {
                        st.textContent = 'Selecting best node…';
                    } else if (idx === 3) {
                        st.textContent = 'Probing RPC…';
                    } else {
                        st.textContent = 'Working…';
                    }
                }
                idx += 1;
                setTimeout(advance, 650);
            } else {
                // simulate node selection from pool, then launch
                setTimeout(() => {
                    if (!bootInProgress) return;
                    openWorkstationPopup();
                    // small delay then hide overlay
                    setTimeout(() => {
                        bootOverlay.classList.remove('boot-overlay--visible');
                        bootOverlay.classList.add('boot-overlay--hidden');
                        bootInProgress = false;
                    }, 600);
                }, 350);
            }
        }
        advance();
    }

    if (launchBtn && bootOverlay) {
        launchBtn.addEventListener('click', (e) => {
            e.preventDefault();
            runBootSequence();
        });
    }

    const secondaryLaunchBtn = document.getElementById('btn-launch-workstation-secondary');
    if (secondaryLaunchBtn && bootOverlay) {
        secondaryLaunchBtn.addEventListener('click', (e) => {
            e.preventDefault();
            runBootSequence();
        });
    }


    if (bootCancel && bootOverlay) {
        bootCancel.addEventListener('click', () => {
            bootInProgress = false;
            bootOverlay.classList.remove('boot-overlay--visible');
            bootOverlay.classList.add('boot-overlay--hidden');
        });
    }


})();
