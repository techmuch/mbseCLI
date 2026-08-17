/**
 * mbseCLI Website Interactive Enhancements
 */

// Model data for the interactive visualizer mockup
const mockModel = {
  'DronePackage': {
    name: 'DronePackage',
    kind: 'package',
    fqn: 'DronePackage',
    file: 'examples/drone.sysml:1',
    attributes: [
      { key: 'version', val: '"2.4.0"' },
      { key: 'standard', val: 'SysML v2 / KerML' }
    ],
    notes: [
      { author: 'alex', status: 'in_review', text: 'Top-level package containing physical and avionics definitions.' }
    ]
  },
  'Quadcopter': {
    name: 'Quadcopter',
    kind: 'part def',
    fqn: 'DronePackage::Quadcopter',
    file: 'examples/drone.sysml:4',
    attributes: [
      { key: 'massLimit', val: '2.5 kg' },
      { key: 'maxThrust', val: '40.0 N' },
      { key: 'endurance', val: '35 min' }
    ],
    notes: [
      { author: 'david', status: 'resolved', text: 'Verified frame structural integrity matches rotor thrust limits.' }
    ]
  },
  'PropulsionSystem': {
    name: 'PropulsionSystem',
    kind: 'part',
    fqn: 'DronePackage::Quadcopter::propulsion',
    file: 'examples/drone.sysml:12',
    attributes: [
      { key: 'motorCount', val: '4 (BLDC)' },
      { key: 'escVoltage', val: '24V' },
      { key: 'propDiameter', val: '10 inch' }
    ],
    notes: [
      { author: 'elena', status: 'open', text: 'Check ESC PWM frequency sync with flight controller.' }
    ]
  },
  'AvionicsBay': {
    name: 'AvionicsBay',
    kind: 'part',
    fqn: 'DronePackage::Quadcopter::avionics',
    file: 'examples/drone.sysml:22',
    attributes: [
      { key: 'cpu', val: 'STM32H7 Dual Core' },
      { key: 'imu', val: 'Dual Redundant 6-DoF' },
      { key: 'telemetryRate', val: '50 Hz' }
    ],
    notes: [
      { author: 'marcus', status: 'in_review', text: 'Telemetry port connection verified against ground station receiver.' }
    ]
  }
};

document.addEventListener('DOMContentLoaded', () => {
  initClipboard();
  initTabs();
  initMockupInteractive();
});

// Copy to Clipboard with visual feedback
function initClipboard() {
  document.querySelectorAll('.copy-btn').forEach(btn => {
    btn.addEventListener('click', async () => {
      const code = btn.getAttribute('data-copy');
      if (!code) return;

      try {
        await navigator.clipboard.writeText(code);
        const originalText = btn.innerHTML;
        btn.classList.add('copied');
        btn.innerHTML = `
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"></polyline></svg>
          <span>Copied!</span>
        `;
        setTimeout(() => {
          btn.classList.remove('copied');
          btn.innerHTML = originalText;
        }, 2000);
      } catch (err) {
        console.error('Failed to copy', err);
      }
    });
  });
}

// Tab Switching
function initTabs() {
  const tabs = document.querySelectorAll('.tab-btn');
  tabs.forEach(tab => {
    tab.addEventListener('click', () => {
      const targetId = tab.getAttribute('data-tab');
      const container = tab.closest('.install-tabs-wrapper');
      if (!container) return;

      container.querySelectorAll('.tab-btn').forEach(t => t.classList.remove('active'));
      container.querySelectorAll('.tab-pane').forEach(p => p.classList.remove('active'));

      tab.classList.add('active');
      const targetPane = container.querySelector(`#${targetId}`);
      if (targetPane) targetPane.classList.add('active');
    });
  });
}

// Interactive Mockup Sync
function initMockupInteractive() {
  const treeNodes = document.querySelectorAll('.tree-node[data-element]');
  const diagramNodes = document.querySelectorAll('.diagram-node[data-element]');
  
  const inspectorName = document.getElementById('inspect-name');
  const inspectorKind = document.getElementById('inspect-kind');
  const inspectorFqn = document.getElementById('inspect-fqn');
  const inspectorFile = document.getElementById('inspect-file');
  const inspectorProps = document.getElementById('inspect-props');
  const inspectorNotes = document.getElementById('inspect-notes');

  function selectElement(name) {
    const data = mockModel[name];
    if (!data) return;

    // Update active tree nodes
    treeNodes.forEach(node => {
      if (node.getAttribute('data-element') === name) {
        node.classList.add('active');
      } else {
        node.classList.remove('active');
      }
    });

    // Update active diagram nodes
    diagramNodes.forEach(node => {
      if (node.getAttribute('data-element') === name) {
        node.classList.add('active');
      } else {
        node.classList.remove('active');
      }
    });

    // Update Inspector
    if (inspectorName) inspectorName.textContent = data.name;
    if (inspectorKind) inspectorKind.textContent = data.kind;
    if (inspectorFqn) inspectorFqn.textContent = data.fqn;
    if (inspectorFile) inspectorFile.textContent = data.file;

    if (inspectorProps) {
      inspectorProps.innerHTML = data.attributes.map(attr => `
        <div class="property-row">
          <span class="property-key">${attr.key}</span>
          <span class="property-val">${attr.val}</span>
        </div>
      `).join('');
    }

    if (inspectorNotes) {
      inspectorNotes.innerHTML = data.notes.map(note => `
        <div class="feedback-note-card">
          <div class="feedback-header">
            <span class="feedback-author">@${note.author}</span>
            <span class="feedback-status">${note.status}</span>
          </div>
          <p class="feedback-text">${note.text}</p>
        </div>
      `).join('');
    }
  }

  treeNodes.forEach(node => {
    node.addEventListener('click', () => {
      const el = node.getAttribute('data-element');
      selectElement(el);
    });
  });

  diagramNodes.forEach(node => {
    node.addEventListener('click', () => {
      const el = node.getAttribute('data-element');
      selectElement(el);
    });
  });
}
