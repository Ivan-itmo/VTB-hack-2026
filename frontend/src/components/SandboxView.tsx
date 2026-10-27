import React, { useState } from 'react';
import {
  GraduationCap,
  Send,
  Sparkles,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  Award,
  FileCheck,
  ChevronDown,
  ChevronUp,
  Copy,
  Check,
  HelpCircle,
} from 'lucide-react';
import type { SandboxCase, SandboxResponse } from '../types';

interface SandboxViewProps {
  taskText: string;
  setTaskText?: (val: string) => void;
  traineeSolution: string;
  setTraineeSolution: (val: string) => void;
  selectedCase: SandboxCase | null;
  onApplySampleAnswer: () => void;
  onSubmit: () => void;
  isLoading: boolean;
  response: SandboxResponse | null;
  error: string | null;
}

export const SandboxView: React.FC<SandboxViewProps> = ({
  taskText,
  traineeSolution,
  setTraineeSolution,
  selectedCase,
  onApplySampleAnswer,
  onSubmit,
  isLoading,
  response,
  error,
}) => {
  const [copiedCorrect, setCopiedCorrect] = useState(false);
  const [isCorrectExpanded, setIsCorrectExpanded] = useState(true);

  const handleCopyCorrect = () => {
    if (!response?.correct_resolution) return;
    navigator.clipboard.writeText(response.correct_resolution);
    setCopiedCorrect(true);
    setTimeout(() => setCopiedCorrect(false), 2000);
  };

  const getScoreStyle = (score: number) => {
    if (score >= 8) {
      return {
        bg: 'bg-emerald-50 border-emerald-300 text-emerald-800',
        badgeBg: 'bg-emerald-600 text-white',
        barColor: 'bg-emerald-500',
        title: 'Успешно пройдено',
        icon: <CheckCircle2 className="w-6 h-6 text-emerald-600" />,
      };
    } else if (score >= 5) {
      return {
        bg: 'bg-amber-50 border-amber-300 text-amber-900',
        badgeBg: 'bg-amber-600 text-white',
        barColor: 'bg-amber-500',
        title: 'Требует доработки',
        icon: <AlertTriangle className="w-6 h-6 text-amber-600" />,
      };
    } else {
      return {
        bg: 'bg-rose-50 border-rose-300 text-rose-900',
        badgeBg: 'bg-rose-600 text-white',
        barColor: 'bg-rose-500',
        title: 'Не пройдено: Нарушение регламентов',
        icon: <XCircle className="w-6 h-6 text-rose-600" />,
      };
    }
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
      {/* LEFT COLUMN: Trainee Inputs */}
      <div className="lg:col-span-5 space-y-4">
        {/* 1. Problem Statement Card */}
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
          <div className="bg-gradient-to-r from-[#002882] to-[#0040A8] px-5 py-3 text-white flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <GraduationCap className="w-4 h-4 text-blue-200" />
              <h3 className="font-semibold text-sm">Учебная задача для стажера</h3>
            </div>
            {selectedCase && (
              <span className="text-[11px] bg-white/20 text-white font-medium px-2 py-0.5 rounded">
                Кейс #{selectedCase.id.slice(-2)}
              </span>
            )}
          </div>

          <div className="p-4 bg-slate-50/70 border-b border-slate-200">
            <label className="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">
              Условие кейса
            </label>
            <p className="text-xs sm:text-sm text-slate-800 leading-relaxed font-sans">
              {taskText || 'Выберите учебный кейс из списка вверху или введите текст задачи вручную.'}
            </p>
          </div>

          {/* 2. Trainee Solution Area */}
          <div className="p-5 space-y-4">
            <div>
              <div className="flex items-center justify-between mb-1.5">
                <label
                  htmlFor="trainee-answer"
                  className="block text-xs font-bold uppercase tracking-wider text-slate-600"
                >
                  Ваш проект решения / Ответ операциониста
                </label>
                <span className="text-[11px] text-slate-400">
                  {traineeSolution.length} симв.
                </span>
              </div>
              <textarea
                id="trainee-answer"
                rows={7}
                value={traineeSolution}
                onChange={(e) => setTraineeSolution(e.target.value)}
                placeholder="Опишите ваши действия и регламентную позицию по инциденту..."
                className="w-full text-sm text-slate-800 bg-slate-50/50 border border-slate-300 rounded-lg p-3.5 focus:outline-none focus:ring-2 focus:ring-[#0050CF] focus:border-transparent transition-all leading-relaxed placeholder:text-slate-400 font-sans"
              />
            </div>

            {/* Action buttons */}
            <div className="flex flex-col sm:flex-row gap-2.5 pt-1">
              <button
                type="button"
                onClick={onApplySampleAnswer}
                className="flex-1 inline-flex items-center justify-center px-4 py-2.5 text-xs font-semibold text-[#002882] bg-blue-50 hover:bg-blue-100/80 border border-blue-200 rounded-lg transition-colors cursor-pointer"
                title="Подставить типичный ответ стажера для проверки анализа ошибок"
              >
                <Sparkles className="w-3.5 h-3.5 mr-1.5 text-[#0050CF]" />
                Подставить типовой ответ
              </button>

              <button
                type="button"
                onClick={onSubmit}
                disabled={isLoading || !traineeSolution.trim()}
                className="flex-1 inline-flex items-center justify-center px-5 py-2.5 text-xs font-semibold text-white bg-[#002882] hover:bg-[#001f66] active:bg-[#00174c] disabled:opacity-50 disabled:cursor-not-allowed rounded-lg shadow-sm transition-all cursor-pointer"
              >
                {isLoading ? (
                  <>
                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <span>Оценка наставника...</span>
                  </>
                ) : (
                  <>
                    <Send className="w-3.5 h-3.5 mr-1.5" />
                    <span>Проверить ответ</span>
                  </>
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Evaluation criteria card */}
        <div className="bg-slate-100/80 rounded-xl p-4 border border-slate-200/80 text-xs text-slate-600 space-y-2">
          <div className="flex items-center space-x-1.5 text-slate-700 font-bold">
            <HelpCircle className="w-4 h-4 text-[#0050CF]" />
            <span>Критерии оценки тренажера</span>
          </div>
          <ul className="list-disc list-inside space-y-1 pl-1">
            <li>Точность соблюдения контрольных процедур (идентификация, кодовое слово).</li>
            <li>Отсутствие обещаний невыполнимых сроков или несанкционированных списаний.</li>
            <li>Правильная маршрутизация инцидента на соответствующую линию поддержки.</li>
          </ul>
        </div>
      </div>

      {/* RIGHT COLUMN: Mentor Feedback and Evaluation */}
      <div className="lg:col-span-7 space-y-5">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 p-4 rounded-xl text-sm flex items-start space-x-3">
            <AlertTriangle className="w-5 h-5 text-red-500 shrink-0 mt-0.5" />
            <div>
              <p className="font-semibold">Ошибка проверки ответа:</p>
              <p className="text-xs text-red-600 mt-1">{error}</p>
            </div>
          </div>
        )}

        {!response && !isLoading && !error && (
          <div className="bg-white rounded-xl shadow-sm border border-dashed border-slate-300 p-12 text-center">
            <div className="w-14 h-14 bg-blue-50 text-[#002882] rounded-full flex items-center justify-center mx-auto mb-4 border border-blue-100">
              <GraduationCap className="w-7 h-7" />
            </div>
            <h4 className="text-base font-semibold text-slate-700 mb-1">
              Режим проверки знаний стажера
            </h4>
            <p className="text-xs text-slate-500 max-w-md mx-auto">
              Напишите ваш вариант решения или воспользуйтесь кнопкой «Подставить типовой ответ», затем нажмите «Проверить ответ» для получения подробного разбора от виртуального наставника.
            </p>
          </div>
        )}

        {isLoading && (
          <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-10 text-center space-y-4">
            <div className="relative w-12 h-12 mx-auto">
              <div className="w-12 h-12 rounded-full border-4 border-blue-100 border-t-[#0050CF] animate-spin"></div>
            </div>
            <p className="text-sm font-semibold text-slate-700">
              Проверка решения по чек-листу наставника...
            </p>
            <p className="text-xs text-slate-400">
              Сравнение с эталоном, поиск критических уязвимостей и нарушений регламентов
            </p>
          </div>
        )}

        {response && !isLoading && (
          <div className="space-y-5 animate-in fade-in duration-300">
            {/* 1. Score Badge & Verdict */}
            {(() => {
              const style = getScoreStyle(response.score);
              return (
                <div className={`rounded-xl p-5 border shadow-xs transition-all ${style.bg}`}>
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                    <div className="flex items-start space-x-3.5">
                      <div className="p-2 rounded-xl bg-white/80 shadow-xs shrink-0">
                        {style.icon}
                      </div>
                      <div>
                        <div className="flex items-center space-x-2">
                          <span className="text-xs font-bold uppercase tracking-wider opacity-80">
                            Вердикт наставника
                          </span>
                          <span
                            className={`px-2.5 py-0.5 rounded-full text-xs font-extrabold ${style.badgeBg}`}
                          >
                            {response.score} / 10 баллов
                          </span>
                        </div>
                        <h4 className="text-base font-bold mt-0.5">{style.title}</h4>
                      </div>
                    </div>

                    {/* Progress score visual */}
                    <div className="bg-white/90 px-4 py-2.5 rounded-lg border border-slate-200/70 min-w-[180px]">
                      <div className="flex justify-between text-xs font-semibold mb-1 text-slate-700">
                        <span>Балл соответствия</span>
                        <span>{response.score * 10}%</span>
                      </div>
                      <div className="w-full bg-slate-200 h-2.5 rounded-full overflow-hidden">
                        <div
                          className={`h-full rounded-full transition-all duration-500 ${style.barColor}`}
                          style={{ width: `${response.score * 10}%` }}
                        ></div>
                      </div>
                    </div>
                  </div>
                </div>
              );
            })()}

            {/* 2. Detailed Feedback Card */}
            <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
              <div className="bg-slate-50/80 px-5 py-3 border-b border-slate-200 flex items-center space-x-2">
                <FileCheck className="w-4 h-4 text-[#002882]" />
                <h4 className="text-xs font-bold uppercase tracking-wider text-slate-700">
                  Разбор ответа и комментарий наставника
                </h4>
              </div>
              <div className="p-5">
                <p className="text-xs sm:text-sm text-slate-800 leading-relaxed font-sans bg-slate-50/70 p-4 rounded-lg border border-slate-200/80">
                  {response.feedback}
                </p>
              </div>
            </div>

            {/* 3. Missed Rules Block */}
            {response.missed_rules && response.missed_rules.length > 0 && (
              <div className="bg-white rounded-xl shadow-sm border border-rose-200 overflow-hidden">
                <div className="bg-rose-50 px-5 py-3 border-b border-rose-200 flex items-center justify-between">
                  <div className="flex items-center space-x-2 text-rose-800">
                    <AlertTriangle className="w-4 h-4 text-rose-600" />
                    <h4 className="text-xs font-bold uppercase tracking-wider">
                      Упущенные регламенты и пункты правил ({response.missed_rules.length})
                    </h4>
                  </div>
                  <span className="text-[11px] font-bold text-rose-700 bg-rose-100 px-2 py-0.5 rounded">
                    Критично
                  </span>
                </div>

                <div className="p-5">
                  <ul className="space-y-2.5">
                    {response.missed_rules.map((rule, idx) => (
                      <li
                        key={idx}
                        className="flex items-start space-x-2.5 text-xs text-slate-800 bg-rose-50/40 p-3 rounded-lg border border-rose-100"
                      >
                        <span className="w-5 h-5 rounded-full bg-rose-200 text-rose-800 font-bold text-[10px] flex items-center justify-center shrink-0 mt-0.5">
                          {idx + 1}
                        </span>
                        <span className="leading-relaxed font-medium">{rule}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            )}

            {/* 4. Correct Resolution (Эталонное решение) */}
            <div className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden">
              <div
                className="bg-slate-50/80 px-5 py-3 border-b border-slate-200 flex items-center justify-between cursor-pointer select-none"
                onClick={() => setIsCorrectExpanded(!isCorrectExpanded)}
              >
                <div className="flex items-center space-x-2">
                  <Award className="w-4 h-4 text-[#0050CF]" />
                  <h4 className="text-xs font-bold uppercase tracking-wider text-slate-800">
                    Эталонное регламентное решение
                  </h4>
                </div>
                <div className="flex items-center space-x-2">
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleCopyCorrect();
                    }}
                    className="inline-flex items-center space-x-1 text-xs font-medium text-slate-600 hover:text-[#0050CF] bg-white hover:bg-blue-50 px-2.5 py-1 rounded border border-slate-200 transition-colors cursor-pointer"
                  >
                    {copiedCorrect ? (
                      <>
                        <Check className="w-3 h-3 text-emerald-600" />
                        <span className="text-emerald-700 text-[11px]">Скопировано</span>
                      </>
                    ) : (
                      <>
                        <Copy className="w-3 h-3" />
                        <span className="text-[11px]">Копировать</span>
                      </>
                    )}
                  </button>
                  {isCorrectExpanded ? (
                    <ChevronUp className="w-4 h-4 text-slate-400" />
                  ) : (
                    <ChevronDown className="w-4 h-4 text-slate-400" />
                  )}
                </div>
              </div>

              {isCorrectExpanded && (
                <div className="p-5">
                  <pre className="text-xs sm:text-sm font-sans text-slate-800 whitespace-pre-wrap leading-relaxed bg-blue-50/40 p-4 rounded-lg border border-blue-100">
                    {response.correct_resolution}
                  </pre>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
