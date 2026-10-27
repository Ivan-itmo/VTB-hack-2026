import React from 'react';
import { Layers, ChevronDown, RefreshCw } from 'lucide-react';
import type { AppMode, CopilotCase, SandboxCase } from '../types';

interface CaseSelectorProps {
  mode: AppMode;
  copilotCases: CopilotCase[];
  sandboxCases: SandboxCase[];
  selectedCaseId: string | null;
  onSelectCopilotCase: (c: CopilotCase) => void;
  onSelectSandboxCase: (c: SandboxCase) => void;
  onReset: () => void;
  isLoadingCases: boolean;
}

export const CaseSelector: React.FC<CaseSelectorProps> = ({
  mode,
  copilotCases,
  sandboxCases,
  selectedCaseId,
  onSelectCopilotCase,
  onSelectSandboxCase,
  onReset,
  isLoadingCases,
}) => {
  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const val = e.target.value;
    if (!val) {
      onReset();
      return;
    }

    if (mode === 'copilot') {
      const found = copilotCases.find((c) => c.id === val);
      if (found) onSelectCopilotCase(found);
    } else {
      const found = sandboxCases.find((c) => c.id === val);
      if (found) onSelectSandboxCase(found);
    }
  };

  return (
    <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-4 mb-6 transition-all">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center space-x-2 text-slate-800">
          <div className="p-2 bg-blue-50 text-[#002882] rounded-lg border border-blue-100">
            <Layers className="w-5 h-5" />
          </div>
          <div>
            <h2 className="text-sm font-bold uppercase tracking-wider text-slate-500">
              Быстрый старт: Демо-кейсы бэк-офиса
            </h2>
            <p className="text-xs text-slate-500">
              Выберите сценарий из базы регламентов для автоматического заполнения
            </p>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <div className="relative min-w-[280px] sm:min-w-[360px]">
            <select
              value={selectedCaseId || ''}
              onChange={handleChange}
              disabled={isLoadingCases}
              className="w-full appearance-none bg-slate-50 border border-slate-300 text-slate-800 text-sm font-medium rounded-lg px-3.5 py-2.5 pr-10 focus:outline-none focus:ring-2 focus:ring-[#0050CF] focus:border-transparent cursor-pointer shadow-inner disabled:opacity-50"
            >
              <option value="">— Выбрать демо-кейс из базы регламентов —</option>
              {mode === 'copilot' ? (
                <>
                  <optgroup label="Регламентные операции и эскалация">
                    {copilotCases.map((c) => (
                      <option key={c.id} value={c.id}>
                        [{c.reg_code}] {c.title}
                      </option>
                    ))}
                  </optgroup>
                </>
              ) : (
                <>
                  <optgroup label="Учебные кейсы для стажеров">
                    {sandboxCases.map((c) => (
                      <option key={c.id} value={c.id}>
                        {c.title}
                      </option>
                    ))}
                  </optgroup>
                </>
              )}
            </select>
            <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-3 text-slate-500">
              <ChevronDown className="w-4 h-4" />
            </div>
          </div>

          {selectedCaseId && (
            <button
              type="button"
              onClick={onReset}
              className="p-2.5 text-slate-500 hover:text-slate-800 hover:bg-slate-100 rounded-lg border border-slate-200 transition-colors text-xs flex items-center space-x-1 cursor-pointer"
              title="Очистить поля"
            >
              <RefreshCw className="w-4 h-4" />
              <span className="hidden md:inline">Сброс</span>
            </button>
          )}
        </div>
      </div>

      {/* Selected case hint banner */}
      {selectedCaseId && (
        <div className="mt-3 pt-3 border-t border-slate-100 flex items-center justify-between text-xs text-slate-600">
          <div className="flex items-center space-x-2">
            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-blue-100 text-[#002882]">
              Кейс загружен
            </span>
            <span className="font-medium text-slate-700">
              Данные кейса подставлены в форму ввода слева.
            </span>
          </div>
          <span className="text-slate-400">
            ID: <code className="font-mono text-slate-600">{selectedCaseId}</code>
          </span>
        </div>
      )}
    </div>
  );
};
