import { useState, useEffect } from 'react';
import { Header } from './components/Header';
import { CaseSelector } from './components/CaseSelector';
import { CopilotView } from './components/CopilotView';
import { SandboxView } from './components/SandboxView';
import type {
  AppMode,
  CopilotCase,
  SandboxCase,
  CasesData,
  CopilotResponse,
  SandboxResponse,
} from './types';

const API_BASE_URL = 'http://localhost:8000';

export function App() {
  const [mode, setMode] = useState<AppMode>('copilot');
  const [isBackendConnected, setIsBackendConnected] = useState<boolean | null>(null);

  // Loaded cases from backend
  const [casesData, setCasesData] = useState<CasesData>({
    copilot_cases: [],
    sandbox_cases: [],
  });
  const [isLoadingCases, setIsLoadingCases] = useState<boolean>(true);

  // Selected case
  const [selectedCaseId, setSelectedCaseId] = useState<string | null>(null);

  // Copilot State
  const [copilotTaskText, setCopilotTaskText] = useState<string>('');
  const [selectedCopilotCase, setSelectedCopilotCase] = useState<CopilotCase | null>(null);
  const [copilotResponse, setCopilotResponse] = useState<CopilotResponse | null>(null);
  const [isCopilotLoading, setIsCopilotLoading] = useState<boolean>(false);
  const [copilotError, setCopilotError] = useState<string | null>(null);

  // Sandbox State
  const [sandboxTaskText, setSandboxTaskText] = useState<string>('');
  const [sandboxTraineeSolution, setSandboxTraineeSolution] = useState<string>('');
  const [selectedSandboxCase, setSelectedSandboxCase] = useState<SandboxCase | null>(null);
  const [sandboxResponse, setSandboxResponse] = useState<SandboxResponse | null>(null);
  const [isSandboxLoading, setIsSandboxLoading] = useState<boolean>(false);
  const [sandboxError, setSandboxError] = useState<string | null>(null);

  // Check backend health and load cases
  useEffect(() => {
    const fetchCases = async () => {
      try {
        setIsLoadingCases(true);
        const res = await fetch(`${API_BASE_URL}/api/v1/cases`);
        if (!res.ok) {
          throw new Error(`HTTP error ${res.status}`);
        }
        const data: CasesData = await res.json();
        setCasesData(data);
        setIsBackendConnected(true);

        // Auto-select first copilot case as initial preview
        if (data.copilot_cases.length > 0) {
          const first = data.copilot_cases[0];
          setSelectedCaseId(first.id);
          setSelectedCopilotCase(first);
          setCopilotTaskText(first.incoming_task);
        }
      } catch (err) {
        console.warn('Backend cases endpoint not reached yet:', err);
        setIsBackendConnected(false);
      } finally {
        setIsLoadingCases(false);
      }
    };

    fetchCases();

    // Periodic ping to keep connection status active
    const interval = setInterval(async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/`, { method: 'GET' });
        setIsBackendConnected(res.ok);
      } catch {
        setIsBackendConnected(false);
      }
    }, 15000);

    return () => clearInterval(interval);
  }, []);

  // Handle switching mode
  const handleModeChange = (newMode: AppMode) => {
    setMode(newMode);
    setSelectedCaseId(null);

    if (newMode === 'copilot') {
      if (casesData.copilot_cases.length > 0) {
        const first = casesData.copilot_cases[0];
        setSelectedCaseId(first.id);
        setSelectedCopilotCase(first);
        setCopilotTaskText(first.incoming_task);
      }
    } else {
      if (casesData.sandbox_cases.length > 0) {
        const first = casesData.sandbox_cases[0];
        setSelectedCaseId(first.id);
        setSelectedSandboxCase(first);
        setSandboxTaskText(first.incoming_task);
        setSandboxTraineeSolution('');
        setSandboxResponse(null);
      }
    }
  };

  // Case Selection Handlers
  const handleSelectCopilotCase = (c: CopilotCase) => {
    setSelectedCaseId(c.id);
    setSelectedCopilotCase(c);
    setCopilotTaskText(c.incoming_task);
    setCopilotResponse(null);
    setCopilotError(null);
  };

  const handleSelectSandboxCase = (c: SandboxCase) => {
    setSelectedCaseId(c.id);
    setSelectedSandboxCase(c);
    setSandboxTaskText(c.incoming_task);
    setSandboxTraineeSolution('');
    setSandboxResponse(null);
    setSandboxError(null);
  };

  const handleReset = () => {
    setSelectedCaseId(null);
    if (mode === 'copilot') {
      setSelectedCopilotCase(null);
      setCopilotTaskText('');
      setCopilotResponse(null);
      setCopilotError(null);
    } else {
      setSelectedSandboxCase(null);
      setSandboxTaskText('');
      setSandboxTraineeSolution('');
      setSandboxResponse(null);
      setSandboxError(null);
    }
  };

  // Preset quick fill
  const handleApplyCopilotSample = () => {
    if (selectedCopilotCase) {
      setCopilotTaskText(selectedCopilotCase.incoming_task);
    } else if (casesData.copilot_cases.length > 0) {
      const first = casesData.copilot_cases[0];
      setSelectedCaseId(first.id);
      setSelectedCopilotCase(first);
      setCopilotTaskText(first.incoming_task);
    }
  };

  const handleApplySandboxSample = () => {
    if (selectedSandboxCase) {
      setSandboxTraineeSolution(selectedSandboxCase.sample_trainee_answer);
    } else if (casesData.sandbox_cases.length > 0) {
      const first = casesData.sandbox_cases[0];
      setSelectedCaseId(first.id);
      setSelectedSandboxCase(first);
      setSandboxTaskText(first.incoming_task);
      setSandboxTraineeSolution(first.sample_trainee_answer);
    }
  };

  // Submissions
  const handleResolveCopilot = async () => {
    if (!copilotTaskText.trim()) return;

    try {
      setIsCopilotLoading(true);
      setCopilotError(null);

      const res = await fetch(`${API_BASE_URL}/api/v1/copilot/resolve`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          task_text: copilotTaskText,
          case_id: selectedCaseId || undefined,
        }),
      });

      if (!res.ok) {
        throw new Error(`Ошибка сервера: ${res.status} ${res.statusText}`);
      }

      const data: CopilotResponse = await res.json();
      setCopilotResponse(data);
    } catch (err: any) {
      console.error(err);
      setCopilotError(
        err.message || 'Не удалось связаться с сервером. Проверьте запуск backend на порту 8000.'
      );
    } finally {
      setIsCopilotLoading(false);
    }
  };

  const handleEvaluateSandbox = async () => {
    if (!sandboxTraineeSolution.trim()) return;

    try {
      setIsSandboxLoading(true);
      setSandboxError(null);

      const res = await fetch(`${API_BASE_URL}/api/v1/sandbox/evaluate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          task_text: sandboxTaskText,
          trainee_solution: sandboxTraineeSolution,
          case_id: selectedCaseId || undefined,
        }),
      });

      if (!res.ok) {
        throw new Error(`Ошибка сервера: ${res.status} ${res.statusText}`);
      }

      const data: SandboxResponse = await res.json();
      setSandboxResponse(data);
    } catch (err: any) {
      console.error(err);
      setSandboxError(
        err.message || 'Не удалось связаться с сервером. Проверьте запуск backend на порту 8000.'
      );
    } finally {
      setIsSandboxLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#F4F6F9] text-slate-900 flex flex-col font-sans">
      {/* Top Header */}
      <Header
        mode={mode}
        onModeChange={handleModeChange}
        isBackendConnected={isBackendConnected}
      />

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {/* Demo Case Selector Bar */}
        <CaseSelector
          mode={mode}
          copilotCases={casesData.copilot_cases}
          sandboxCases={casesData.sandbox_cases}
          selectedCaseId={selectedCaseId}
          onSelectCopilotCase={handleSelectCopilotCase}
          onSelectSandboxCase={handleSelectSandboxCase}
          onReset={handleReset}
          isLoadingCases={isLoadingCases}
        />

        {/* Active Mode View */}
        {mode === 'copilot' ? (
          <CopilotView
            taskText={copilotTaskText}
            setTaskText={setCopilotTaskText}
            selectedCase={selectedCopilotCase}
            onApplySampleText={handleApplyCopilotSample}
            onSubmit={handleResolveCopilot}
            isLoading={isCopilotLoading}
            response={copilotResponse}
            error={copilotError}
          />
        ) : (
          <SandboxView
            taskText={sandboxTaskText}
            traineeSolution={sandboxTraineeSolution}
            setTraineeSolution={setSandboxTraineeSolution}
            selectedCase={selectedSandboxCase}
            onApplySampleAnswer={handleApplySandboxSample}
            onSubmit={handleEvaluateSandbox}
            isLoading={isSandboxLoading}
            response={sandboxResponse}
            error={sandboxError}
          />
        )}
      </main>

      {/* Footer */}
      <footer className="bg-white border-t border-slate-200 py-4 mt-auto">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col sm:flex-row items-center justify-between text-xs text-slate-500 gap-2">
          <div className="flex items-center space-x-2">
            <span className="font-semibold text-slate-700">ВТБ Ко-пилот</span>
            <span>•</span>
            <span>Система поддержки принятия решений бэк-офиса</span>
          </div>
          <div className="flex items-center space-x-4">
            <span>FastAPI (Port 8000)</span>
            <span>•</span>
            <span>React + Vite (Port 5173)</span>
            <span>•</span>
            <span className="text-[#0050CF] font-medium">Strict Grounding Active</span>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default App;
