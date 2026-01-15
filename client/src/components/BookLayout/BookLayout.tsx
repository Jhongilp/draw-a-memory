import { Outlet } from 'react-router-dom';
import { PageSidebar } from '../PageSidebar';
import type { PageDraft } from '../../types/photo';

interface BookLayoutProps {
  pages: PageDraft[];
  onReorderPages?: (pages: PageDraft[]) => void;
}

export function BookLayout({ pages, onReorderPages }: BookLayoutProps) {
  return (
    <div className="flex flex-1 overflow-hidden">
      <PageSidebar pages={pages} onReorder={onReorderPages} />
      <Outlet />
    </div>
  );
}
