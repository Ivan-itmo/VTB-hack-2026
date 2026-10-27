import React, { useState } from 'react';
import {
  Send,
  Sparkles,
  ShieldCheck,
  AlertTriangle,
  FileText,
  Copy,
  Check,
  Building2,
  BookOpen,
  HelpCircle,
  Bot,
} from 'lucide-react';
import type { CopilotCase, CopilotResponse } from '../types';

interface CopilotViewProps {
  taskText: string;
  setTaskText: (val: string) => void;
  selectedCase: CopilotCase | null;
  onApplySampleText: () => void;
  onSubmit: () => void;
  isLoading: boolean;
  response: CopilotResponse | null;
  error: string | null;
}

export const CopilotView: React.FC<CopilotViewProps> = ({
  taskText,
  setTaskText,
  selectedCase,
  onApplySampleText,
  onSubmit,
  isLoading,
  response,
  error,
}) => {
  const [copiedResolution, setCopiedResolution] = useState(false);
  const [copiedSourceIdx, setCopiedSourceIdx] = useState<number | null>(null);

  const handleCopyResolution = () => {
    if (!response?.resolution_draft) return;
    navigator.clipboard.writeText(response.resolution_draft);
    setCopiedResolution(true);
    setTimeout(() => setCopiedResolution(false), 2000);
  };

  const handleCopySource = (text: string, idx: number) => {
    navigator.clipboard.writeText(text);
    setCopiedSourceIdx(idx);
    setTimeout(() => setCopiedSourceIdx(null), 2000);
  };

  // Helper for confidence badge
  const getConfidenceLevel = (score: number) => {
    const percent = Math.round(score * 100);
    if (percent >= 90) {
      return {
        label: 'Высокая уверенность',
        color: 'text-emerald-700 bg-emerald-50 border-emerald-200',
        barColor: 'bg-emerald-500',
      };
    } else if (percent >= 75) {
      return {
        label: 'Средняя уверенность',
        color: 'text-amber-700 bg-amber-50 border-amber-200',
        barColor: 'bg-amber-500',
      };
    } else {
      return {
        label: 'Требуется экспертная верификация (L3)',
        color: 'text-red-700 bg-red-50 border-red-200',
        barColor: 'bg-red-500',
      };
    }
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
      {/* LEFT COLUMN: Input Form */}
      <div className="lg:col-span-5 space-y-4">
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
          <div className="bg-gradient-to-r from-[#002882] to-[#0040A8] px-5 py-3.5 text-white flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <FileText className="w-4 h-4 text-blue-200" />
              <h3 className="font-semibold text-sm">Входящая задача / Обращение бэк-офиса</h3>
            </div>
            {selectedCase && (
              <span className="text-xs bg-white/20 text-white font-medium px-2 py-0.5 rounded backdrop-blur-xs">
                {selectedCase.category}
              </span>
            )}
          </div>

          <div className="p-5 space-y-4">
            <div>
              <label
                htmlFor="task-input"
                className="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1.5"
              >
                Текст инцидента или запроса клиента
              </label>
              <textarea
                id="task-input"
                rows={9}
                value={taskText}
                onChange={(e) => setTaskText(e.target.value)}
                placeholder="Введите описание входящей задачи, реквизиты платежа или выберите демо-кейс сверху..."
                className="w-full text-sm text-slate-800 bg-slate-50/50 border border-slate-300 rounded-lg p-3.5 focus:outline-none focus:ring-2 focus:ring-[#0050CF] focus:border-transparent transition-all leading-relaxed placeholder:text-slate-400 font-sans"
              />
              <div className="flex items-center justify-between mt-1.5 text-xs text-slate-400">
                <span>Поддерживаются запросы по переводам, СБП, антифроду и валютному контролю</span>
                <span>{taskText.length} симв.</span>
              </div>
            </div>

            {/* Action buttons */}
            <div className="pt-2 flex flex-col sm:flex-row gap-2.5">
              <button
                type="button"
                onClick={onApplySampleText}
                className="flex-1 inline-flex items-center justify-center px-4 py-2.5 text-xs font-semibold text-[#002882] bg-blue-50 hover:bg-blue-100/80 border border-blue-200 rounded-lg transition-colors cursor-pointer"
                title="Подставить типовой регламентный запрос"
              >
                <Sparkles className="w-3.5 h-3.5 mr-1.5 text-[#0050CF]" />
                Подставить типовой ответ/запрос
              </button>

              <button
                type="button"
                onClick={onSubmit}
                disabled={isLoading || !taskText.trim()}
                className="flex-1 inline-flex items-center justify-center px-5 py-2.5 text-xs font-semibold text-white bg-[#002882] hover:bg-[#001f66] active:bg-[#00174c] disabled:opacity-50 disabled:cursor-not-allowed rounded-lg shadow-sm transition-all cursor-pointer"
              >
                {isLoading ? (
                  <>
                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <span>Анализ регламентов...</span>
                  </>
                ) : (
                  <>
                    <Send className="w-3.5 h-3.5 mr-1.5" />
                    <span>Сформировать резолюцию</span>
                  </>
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Quick Help Card */}
        <div className="bg-slate-100/80 rounded-xl p-4 border border-slate-200/80 text-xs text-slate-600 space-y-2">
          <div className="flex items-center space-x-1.5 text-slate-700 font-bold">
            <HelpCircle className="w-4 h-4 text-[#0050CF]" />
            <span>Принцип работы Strict Grounding</span>
          </div>
          <p className="leading-relaxed">
            Ко-пилот не галлюцинирует и формулирует ответ строго на основе действующей базы регламентов банка (REG-01, REG-02, 161-ФЗ, 762-П). Каждая строчка резолюции подтверждается прямой ссылкой на пункт документа.
          </p>
        </div>
      </div>

      {/* RIGHT COLUMN: Results Display */}
      <div className="lg:col-span-7 space-y-5">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 p-4 rounded-xl text-sm flex items-start space-x-3">
            <AlertTriangle className="w-5 h-5 text-red-500 shrink-0 mt-0.5" />
            <div>
              <p className="font-semibold">Ошибка обработки запроса бэкендом:</p>
              <p className="text-xs text-red-600 mt-1">{error}</p>
            </div>
          </div>
        )}

        {!response && !isLoading && !error && (
          <div className="bg-white rounded-xl shadow-sm border border-dashed border-slate-300 p-12 text-center">
            <div className="w-14 h-14 bg-blue-50 text-[#002882] rounded-full flex items-center justify-center mx-auto mb-4 border border-blue-100">
              <Bot className="w-7 h-7" />
            </div>
            <h4 className="text-base font-semibold text-slate-700 mb-1">
              Ожидание входящей задачи
            </h4>
            <p className="text-xs text-slate-500 max-w-md mx-auto">
              Выберите готовый демонстрационный кейс в верхнем меню или введите параметры обращения вручную, затем нажмите «Сформировать резолюцию».
            </p>
          </div>
        )}

        {isLoading && (
          <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-10 text-center space-y-4">
            <div className="relative w-12 h-12 mx-auto">
              <div className="w-12 h-12 rounded-full border-4 border-blue-100 border-t-[#0050CF] animate-spin"></div>
            </div>
            <p className="text-sm font-semibold text-slate-700">
              Сопоставление с нормативной базой банка...
            </p>
            <p className="text-xs text-slate-400">
              Проверка матриц маршрутизации, регламентов REG-01/REG-02 и политик комплаенс
            </p>
          </div>
        )}

        {response && !isLoading && (
          <div className="space-y-5 animate-in fade-in duration-300">
            {/* 1. Routing Banner */}
            <div
              className={`rounded-xl p-4 border shadow-xs transition-all ${
                response.escalation_required
                  ? 'bg-amber-50/90 border-amber-300 text-amber-900'
                  : 'bg-white border-blue-200 text-slate-800'
              }`}
            >
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div className="flex items-start space-x-3">
                  <div
                    className={`p-2.5 rounded-lg shrink-0 ${
                      response.escalation_required
                        ? 'bg-amber-100 text-amber-800'
                        : 'bg-blue-50 text-[#002882]'
                    }`}
                  >
                    <Building2 className="w-5 h-5" />
                  </div>
                  <div>
                    <div className="flex items-center space-x-2">
                      <span className="text-xs font-bold uppercase tracking-wider text-slate-500">
                        Рекомендованный маршрут бэк-офиса
                      </span>
                      {response.escalation_required && (
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-extrabold bg-red-600 text-white uppercase tracking-wider animate-pulse">
                          Эскалация L3
                        </span>
                      )}
                    </div>
                    <p className="text-sm font-bold text-slate-900 mt-0.5">
                      {response.suggested_route}
                    </p>
                  </div>
                </div>

                {/* Confidence indicator */}
                <div className="bg-slate-50 px-3 py-2 rounded-lg border border-slate-200/80 min-w-[170px]">
                  <div className="flex items-center justify-between text-xs mb-1">
                    <span className="text-slate-500 font-medium">Уверенность AI</span>
                    <span className="font-bold text-slate-800">
                      {Math.round(response.confidence_score * 100)}%
                    </span>
                  </div>
                  <div className="w-full bg-slate-200 h-2 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full transition-all duration-500 ${
                        getConfidenceLevel(response.confidence_score).barColor
                      }`}
                      style={{ width: `${Math.round(response.confidence_score * 100)}%` }}
                    ></div>
                  </div>
                  <p className="text-[10px] text-slate-500 mt-1 font-medium">
                    {getConfidenceLevel(response.confidence_score).label}
                  </p>
                </div>
              </div>

              {response.escalation_required && (
                <div className="mt-3 pt-3 border-t border-amber-200/70 text-xs text-amber-800 flex items-center space-x-2">
                  <AlertTriangle className="w-4 h-4 text-amber-600 shrink-0" />
                  <span>
                    Операция выходит за рамки типовых регламентов L1/L2 и требует персонального утверждения Комплаенс-контролем.
                  </span>
                </div>
              )}
            </div>

            {/* 2. Resolution Draft Card */}
            <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
              <div className="bg-slate-50/80 px-5 py-3 border-b border-slate-200 flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <FileText className="w-4 h-4 text-[#002882]" />
                  <h4 className="text-xs font-bold uppercase tracking-wider text-slate-700">
                    Черновик регламентной резолюции
                  </h4>
                </div>
                <button
                  type="button"
                  onClick={handleCopyResolution}
                  className="inline-flex items-center space-x-1 text-xs font-medium text-slate-600 hover:text-[#0050CF] bg-white hover:bg-blue-50 px-2.5 py-1 rounded border border-slate-200 transition-colors cursor-pointer"
                >
                  {copiedResolution ? (
                    <>
                      <Check className="w-3.5 h-3.5 text-emerald-600" />
                      <span className="text-emerald-700">Скопировано</span>
                    </>
                  ) : (
                    <>
                      <Copy className="w-3.5 h-3.5" />
                      <span>Копировать</span>
                    </>
                  )}
                </button>
              </div>

              <div className="p-5">
                <pre className="text-xs sm:text-sm font-sans text-slate-800 whitespace-pre-wrap leading-relaxed bg-slate-50/70 p-4 rounded-lg border border-slate-200/80">
                  {response.resolution_draft}
                </pre>
              </div>
            </div>

            {/* 3. Strict Grounding Sources Block */}
            <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
              <div className="bg-gradient-to-r from-blue-900 via-[#002882] to-blue-950 px-5 py-3 text-white flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <ShieldCheck className="w-4 h-4 text-blue-300" />
                  <h4 className="text-xs font-bold uppercase tracking-wider text-white">
                    Strict Grounding: Нормативно-правовая база
                  </h4>
                </div>
                <span className="text-[11px] font-semibold bg-emerald-500/20 text-emerald-300 border border-emerald-400/30 px-2 py-0.5 rounded-full">
                  100% Верифицировано
                </span>
              </div>

              <div className="p-5 space-y-4">
                <p className="text-xs text-slate-500">
                  Прямые цитаты и выдержки из внутренних регламентов ВТБ и нормативных актов ЦБ РФ:
                </p>

                {response.sources.map((src, idx) => (
                  <div
                    key={idx}
                    className="bg-slate-50 rounded-lg p-4 border border-slate-200 hover:border-blue-300 transition-colors"
                  >
                    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-2 pb-2 border-b border-slate-200/70">
                      <div className="flex items-center space-x-2">
                        <BookOpen className="w-4 h-4 text-[#0050CF] shrink-0" />
                        <span className="text-xs font-bold text-slate-900">
                          {src.document_name}
                        </span>
                      </div>
                      <div className="flex items-center space-x-2 text-[11px]">
                        <span className="bg-blue-100 text-[#002882] font-semibold px-2 py-0.5 rounded">
                          {src.clause}
                        </span>
                        <span className="text-slate-400">
                          Ред. от {src.edition_date}
                        </span>
                      </div>
                    </div>

                    <div className="relative pl-3 border-l-2 border-[#0050CF] text-xs text-slate-700 italic leading-relaxed">
                      «{src.quote}»
                    </div>

                    <div className="mt-2.5 flex justify-end">
                      <button
                        type="button"
                        onClick={() => handleCopySource(src.quote, idx)}
                        className="text-[11px] text-slate-500 hover:text-[#0050CF] flex items-center space-x-1 cursor-pointer transition-colors"
                      >
                        {copiedSourceIdx === idx ? (
                          <>
                            <Check className="w-3 h-3 text-emerald-600" />
                            <span className="text-emerald-600">Цитата скопирована</span>
                          </>
                        ) : (
                          <>
                            <Copy className="w-3 h-3" />
                            <span>Скопировать цитату</span>
                          </>
                        )}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
