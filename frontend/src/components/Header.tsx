import React from 'react';
import { GraduationCap, Bot, Wifi, WifiOff } from 'lucide-react';
import type { AppMode } from '../types';

interface HeaderProps {
  mode: AppMode;
  onModeChange: (mode: AppMode) => void;
  isBackendConnected: boolean | null;
}

export const Header: React.FC<HeaderProps> = ({
  mode,
  onModeChange,
  isBackendConnected,
}) => {
  return (
    <header className="bg-[#002882] text-white shadow-md border-b border-[#001c5c] sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo & Project Title */}
          <div className="flex items-center space-x-3">
            {/* VTB-inspired dynamic emblem */}
            <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-[#0050CF] to-[#001f66] flex items-center justify-center shadow-inner border border-blue-400/30">
              <div className="flex flex-col space-y-0.5 items-center">
                <div className="w-6 h-1 bg-white rounded-full"></div>
                <div className="w-5 h-1 bg-blue-200 rounded-full"></div>
                <div className="w-4 h-1 bg-blue-400 rounded-full"></div>
              </div>
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <span className="font-extrabold text-xl tracking-tight text-white">ВТБ Ко-пилот</span>
                <span className="bg-[#0050CF] text-xs font-semibold px-2 py-0.5 rounded-full text-blue-100 border border-blue-300/30">
                  MVP Спринт №1
                </span>
              </div>
              <p className="text-xs text-blue-200 font-medium">
                Интеллектуальный ассистент и тренажер бэк-офиса
              </p>
            </div>
          </div>

          {/* Mode Switcher */}
          <div className="flex items-center space-x-4">
            <div className="bg-[#001c5c] p-1 rounded-xl flex items-center border border-blue-400/20 shadow-inner">
              <button
                type="button"
                onClick={() => onModeChange('copilot')}
                className={`flex items-center space-x-2 px-4 py-2 rounded-lg text-sm font-semibold transition-all duration-150 cursor-pointer ${
                  mode === 'copilot'
                    ? 'bg-[#0050CF] text-white shadow-md shadow-blue-900/50'
                    : 'text-blue-200 hover:text-white hover:bg-white/5'
                }`}
              >
                <Bot className="w-4 h-4" />
                <span>Боевой режим (Ко-пилот)</span>
              </button>
              <button
                type="button"
                onClick={() => onModeChange('sandbox')}
                className={`flex items-center space-x-2 px-4 py-2 rounded-lg text-sm font-semibold transition-all duration-150 cursor-pointer ${
                  mode === 'sandbox'
                    ? 'bg-[#0050CF] text-white shadow-md shadow-blue-900/50'
                    : 'text-blue-200 hover:text-white hover:bg-white/5'
                }`}
              >
                <GraduationCap className="w-4 h-4" />
                <span>Песочница стажера</span>
              </button>
            </div>

            {/* Backend connectivity indicator */}
            <div
              className={`hidden sm:flex items-center space-x-1.5 px-2.5 py-1 rounded-full text-xs font-medium border ${
                isBackendConnected === true
                  ? 'bg-emerald-950/60 text-emerald-300 border-emerald-500/30'
                  : isBackendConnected === false
                  ? 'bg-red-950/60 text-red-300 border-red-500/30'
                  : 'bg-slate-800 text-slate-300 border-slate-700'
              }`}
              title={
                isBackendConnected === true
                  ? 'Бэкенд доступен (http://localhost:8000)'
                  : 'Бэкенд недоступен или проверяется'
              }
            >
              {isBackendConnected === true ? (
                <>
                  <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
                  <Wifi className="w-3 h-3 text-emerald-400" />
                  <span>Бэкенд 8000</span>
                </>
              ) : (
                <>
                  <span className="w-2 h-2 rounded-full bg-red-400"></span>
                  <WifiOff className="w-3 h-3 text-red-400" />
                  <span>Офлайн</span>
                </>
              )}
            </div>
          </div>
        </div>
      </div>
    </header>
  );
};
