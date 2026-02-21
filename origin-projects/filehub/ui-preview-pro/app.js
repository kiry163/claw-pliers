document.addEventListener('DOMContentLoaded', () => {
  lucide.createIcons();

  // 元素
  const views = document.querySelectorAll('.view');
  const navButtons = document.querySelectorAll('[data-view]');
  const sidebar = document.querySelector('#main-sidebar');
  const sidebarOpen = document.querySelector('#sidebar-open');
  const sidebarClose = document.querySelector('#sidebar-close');
  const sidebarBackdrop = document.querySelector('#sidebar-backdrop');
  
  const drawer = document.querySelector('#task-drawer');
  const drawerBackdrop = document.querySelector('#drawer-backdrop');
  const drawerOpenBtns = document.querySelectorAll('[data-drawer-open]');
  const drawerCloseBtns = document.querySelectorAll('[data-drawer-close]');

  const modalFolder = document.querySelector('#modal-new-folder');
  const btnNewFolder = document.querySelector('#btn-new-folder');
  const closeFolderModal = document.querySelector('#close-folder-modal');

  const dropdownMore = document.querySelector('#dropdown-more');
  const layoutBtns = document.querySelectorAll('[data-layout-btn]');
  const layoutList = document.querySelector('#layout-list');
  const layoutGrid = document.querySelector('#layout-grid');

  const previewTabs = document.querySelectorAll('.tab-preview');
  const panePreview = document.querySelector('#pane-preview');
  const paneMetadata = document.querySelector('#pane-metadata');

  // 视图切换
  const setActiveView = (viewId) => {
    views.forEach(v => v.classList.remove('active'));
    const target = document.querySelector(`#view-${viewId}`);
    if (target) target.classList.add('active');

    navButtons.forEach(btn => btn.classList.toggle('active', btn.dataset.view === viewId));
    
    // 特殊处理登录居中
    if (viewId === 'login') {
      target.classList.add('flex');
    } else {
      document.querySelector('#view-login').classList.remove('flex');
    }

    closeSidebar();
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  navButtons.forEach(btn => btn.addEventListener('click', () => setActiveView(btn.dataset.view)));

  // 布局切换
  layoutBtns.forEach(btn => {
    btn.addEventListener('click', () => {
      const layout = btn.dataset.layoutBtn;
      layoutList.classList.toggle('hidden', layout === 'grid');
      layoutGrid.classList.toggle('hidden', layout === 'list');
      
      layoutBtns.forEach(b => {
        const isActive = b.dataset.layoutBtn === layout;
        b.classList.toggle('bg-white', isActive);
        b.classList.toggle('shadow-sm', isActive);
        b.classList.toggle('text-primary', isActive);
        b.classList.toggle('text-primary/40', !isActive);
      });
    });
  });

  // 模态框逻辑
  btnNewFolder.addEventListener('click', () => modalFolder.classList.replace('hidden', 'flex'));
  closeFolderModal.addEventListener('click', () => modalFolder.classList.replace('flex', 'hidden'));
  modalFolder.addEventListener('click', (e) => { if (e.target === modalFolder) modalFolder.classList.replace('flex', 'hidden'); });

  // 更多操作下拉菜单
  document.querySelectorAll('.btn-more').forEach(btn => {
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      const rect = btn.getBoundingClientRect();
      dropdownMore.style.top = `${rect.bottom + 10}px`;
      dropdownMore.style.left = `${rect.right - 192}px`;
      dropdownMore.classList.remove('hidden');
    });
  });

  window.addEventListener('click', () => dropdownMore.classList.add('hidden'));

  // 预览 Tab 切换
  previewTabs.forEach(tab => {
    tab.addEventListener('click', () => {
      const isPreview = tab.dataset.tab === 'preview';
      panePreview.classList.toggle('hidden', !isPreview);
      paneMetadata.classList.toggle('hidden', isPreview);
      
      previewTabs.forEach(t => {
        const active = t.dataset.tab === tab.dataset.tab;
        t.classList.toggle('bg-white', active);
        t.classList.toggle('shadow-md', active);
        t.classList.toggle('text-primary', active);
        t.classList.toggle('text-primary/40', !active);
      });
    });
  });

  // 侧边栏与抽屉
  const openSidebar = () => { sidebar.classList.add('open'); sidebarBackdrop.classList.add('open'); };
  const closeSidebar = () => { sidebar.classList.remove('open'); sidebarBackdrop.classList.remove('open'); };
  if (sidebarOpen) sidebarOpen.addEventListener('click', openSidebar);
  if (sidebarClose) sidebarClose.addEventListener('click', closeSidebar);
  if (sidebarBackdrop) sidebarBackdrop.addEventListener('click', closeSidebar);

  const openDrawer = () => { drawer.classList.add('open'); drawerBackdrop.classList.add('open'); };
  const closeDrawer = () => { drawer.classList.remove('open'); drawerBackdrop.classList.remove('open'); };
  drawerOpenBtns.forEach(btn => btn.addEventListener('click', openDrawer));
  drawerCloseBtns.forEach(btn => btn.addEventListener('click', closeDrawer));
});
