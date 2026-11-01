import React, { useEffect, useState } from 'react';
import { X, BookOpen, FileText, ShieldAlert, CheckCircle2 } from 'lucide-react';
import type { RuleChunk } from '../types';

interface RuleDrawerProps {
  ruleIdOrCitation: string | null;
  onClose: () => void;
}

export const RuleDrawer: React.FC<RuleDrawerProps> = ({ ruleIdOrCitation, onClose }) => {
  const [rule, setRule] = useState<RuleChunk | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!ruleIdOrCitation) {
      setRule(null);
      return;
    }

    const fetchRule = async () => {
      try {
        setIsLoading(true);
        setError(null);
        // Encode the parameter safely for URL
        const encoded = encodeURIComponent(ruleIdOrCitation);
        const res = await fetch(`/api/v1/rules/${encoded}`);
        if (!res.ok) {
          throw new Error(`Регламент «${ruleIdOrCitation}» не найден в нормативной базе`);
        }
        const data: RuleChunk = await res.json();
        setRule(data);
      } catch (err: any) {
        setError(err.message || 'Ошибка загрузки регламента');
      } finally {
        setIsLoading(false);
      }
    };

    fetchRule();
  }, [ruleIdOrCitation]);

  // Handle ESC key to close drawer
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  if (!ruleIdOrCitation) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-hidden">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-slate-900/50 backdrop-blur-xs transition-opacity animate-in fade-in duration-200"
        onClick={onClose}
      />

      <div className="fixed inset-y-0 right-0 max-w-full flex pl-10">
        <div className="w-screen max-w-md sm:max-w-lg bg-white shadow-2xl border-l border-slate-200 flex flex-col transform transition-transform animate-in slide-in-from-right duration-300">
          {/* Header */}
          <div className="bg-gradient-to-r from-[#002882] to-[#0040A8] px-6 py-4 text-white flex items-center justify-between shadow-md">
            <div className="flex items-center space-x-2.5">
              <div className="p-1.5 bg-white/10 rounded-lg">
                <BookOpen className="w-5 h-5 text-blue-200" />
              </div>
              <div>
                <h3 className="font-bold text-sm tracking-wide">
                  Нормативная база ВТБ
                </h3>
                <p className="text-xs text-blue-100">
                  Strict Grounding: карточка регламента
                </p>
              </div>
            </div>
            <button
              type="button"
              onClick={onClose}
              className="p-1.5 rounded-lg text-blue-100 hover:text-white hover:bg-white/10 transition-colors cursor-pointer"
              title="Закрыть (Esc)"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto p-6 space-y-6">
            {isLoading && (
              <div className="py-16 text-center space-y-4">
                <div className="w-10 h-10 rounded-full border-4 border-blue-100 border-t-[#0050CF] animate-spin mx-auto" />
                <p className="text-sm font-medium text-slate-600">
                  Загрузка пункта из нормативной базы...
                </p>
              </div>
            )}

            {error && !isLoading && (
              <div className="bg-rose-50 border border-rose-200 text-rose-800 p-4 rounded-xl text-sm flex items-start space-x-3">
                <ShieldAlert className="w-5 h-5 text-rose-600 shrink-0 mt-0.5" />
                <div>
                  <p className="font-bold">Пункт не найден</p>
                  <p className="text-xs text-rose-700 mt-1">{error}</p>
                </div>
              </div>
            )}

            {rule && !isLoading && (
              <div className="space-y-5 animate-in fade-in duration-200">
                {/* Meta details */}
                <div className="bg-slate-50 rounded-xl p-4 border border-slate-200 space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-bold uppercase tracking-wider text-slate-500">
                      Идентификатор
                    </span>
                    <span className="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-mono font-bold bg-[#002882] text-white">
                      {rule.id}
                    </span>
                  </div>

                  <div className="border-t border-slate-200/80 pt-2.5 flex items-start justify-between gap-2">
                    <span className="text-xs text-slate-500 font-medium">Документ:</span>
                    <span className="text-xs font-semibold text-slate-800 text-right">
                      {rule.document_title} ({rule.document})
                    </span>
                  </div>

                  {rule.section_title && (
                    <div className="border-t border-slate-200/80 pt-2.5 flex items-start justify-between gap-2">
                      <span className="text-xs text-slate-500 font-medium">Раздел:</span>
                      <span className="text-xs font-medium text-slate-700 text-right">
                        {rule.section}. {rule.section_title}
                      </span>
                    </div>
                  )}

                  <div className="border-t border-slate-200/80 pt-2.5 flex items-center justify-between">
                    <span className="text-xs text-slate-500 font-medium">Официальная цитата:</span>
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-bold bg-blue-100 text-[#002882]">
                      {rule.citation}
                    </span>
                  </div>
                </div>

                {/* Full regulation text */}
                <div className="space-y-2">
                  <div className="flex items-center space-x-1.5 text-xs font-bold uppercase tracking-wider text-slate-700">
                    <FileText className="w-4 h-4 text-[#0050CF]" />
                    <span>Текст нормативного требования</span>
                  </div>
                  <div className="bg-blue-50/50 rounded-xl p-4 border-l-4 border-[#0050CF] border border-blue-100/80 text-sm text-slate-800 leading-relaxed font-sans shadow-xs">
                    {rule.text}
                  </div>
                </div>

                {/* Grounding guarantee */}
                <div className="bg-emerald-50 rounded-xl p-3.5 border border-emerald-200 text-xs text-emerald-800 flex items-start space-x-2.5">
                  <CheckCircle2 className="w-4 h-4 text-emerald-600 shrink-0 mt-0.5" />
                  <div className="leading-relaxed">
                    <span className="font-bold">Strict Grounding верификация: </span>
                    Этот фрагмент нормативной базы используется алгоритмом поиска как единственный разрешенный источник норм. Ко-пилот не формулирует выводов за рамками этого текста.
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Footer */}
          <div className="p-4 border-t border-slate-200 bg-slate-50 flex justify-end">
            <button
              type="button"
              onClick={onClose}
              className="px-5 py-2 text-xs font-semibold text-slate-700 bg-white hover:bg-slate-100 border border-slate-300 rounded-lg shadow-xs transition-colors cursor-pointer"
            >
              Закрыть
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
