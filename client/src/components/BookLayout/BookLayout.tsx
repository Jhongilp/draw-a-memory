import { Outlet } from 'react-router-dom';
import { PageSidebar } from '../PageSidebar';
import type { PageDraft } from '../../types/photo';
import { useAppSelector } from '../../store/hooks';

interface BookLayoutProps {
  onReorderPages?: (pages: PageDraft[]) => void;
}

export function BookLayout({ onReorderPages }: BookLayoutProps) {
  const pages = useAppSelector((state) => state.pages.pages);
  
  return (
    <div className="flex flex-1 overflow-hidden">
      <PageSidebar pages={pages} onReorder={onReorderPages} />
      <Outlet />
    </div>
  );
}
