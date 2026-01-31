import { Outlet } from 'react-router-dom';
import { PageSidebar } from '../PageSidebar';
import type { PageDraft } from '../../types/photo';
import { useAppDispatch, useAppSelector } from '../../store/hooks';
import { reorderPages } from '../../store/slices';

export function BookLayout() {
  const pages = useAppSelector((state) => state.pages.pages);
  const dispatch = useAppDispatch();
    const handleReorderPages = (reorderedPages: PageDraft[]) => {
    dispatch(reorderPages(reorderedPages));
    // TODO: Persist order to server
  };
  
  return (
    <div className="flex flex-1 overflow-hidden">
      <PageSidebar pages={pages} onReorder={handleReorderPages} />
      <Outlet />
    </div>
  );
}
