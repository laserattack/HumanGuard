import { ChangeEvent, useEffect, useMemo, useState } from 'react';
import { AxiosError } from 'axios';
import { FileCategory, getFiles, ManagedFile, uploadFile } from '@/api/files';

type LocalFile = ManagedFile & {
  source: 'server' | 'local-stub';
  localObjectUrl?: string;
};

const categories: Array<{ label: string; value: FileCategory }> = [
  { label: 'Sites', value: 'sites' },
  { label: 'Profiles', value: 'profiles' },
  { label: 'Users', value: 'users' }
];

const formatBytes = (bytes: number) => {
  if (bytes === 0) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / 1024 ** i).toFixed(i > 1 ? 2 : 0)} ${units[i]}`;
};

const parseError = (error: unknown) => {
  const err = error as AxiosError<{ error?: string }>;
  return err.response?.data?.error ?? err.message ?? 'Unknown error';
};

const isUnsupportedEndpoint = (error: unknown) => {
  const err = error as AxiosError;
  return err.response?.status === 404 || err.response?.status === 405;
};

export const FilesManager = () => {
  const [activeCategory, setActiveCategory] = useState<FileCategory>('sites');
  const [filesByCategory, setFilesByCategory] = useState<Record<FileCategory, LocalFile[]>>({
    sites: [],
    profiles: [],
    users: []
  });
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [backendUnsupported, setBackendUnsupported] = useState(false);

  const files = useMemo(() => filesByCategory[activeCategory], [activeCategory, filesByCategory]);

  const loadFiles = async () => {
    if (backendUnsupported) {
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const data = await getFiles(activeCategory);
      setFilesByCategory((prev) => ({
        ...prev,
        [activeCategory]: data.map((item) => ({ ...item, source: 'server' as const }))
      }));
    } catch (e) {
      if (isUnsupportedEndpoint(e)) {
        setBackendUnsupported(true);
        setError(
          'Бэкенд пока не поддерживает файловый API (/api/files, /api/files/upload). Включена визуальная заглушка на фронте.'
        );
      } else {
        setError(parseError(e));
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadFiles();
  }, [activeCategory]);

  useEffect(
    () => () => {
      Object.values(filesByCategory)
        .flat()
        .forEach((file) => {
          if (file.localObjectUrl) {
            URL.revokeObjectURL(file.localObjectUrl);
          }
        });
    },
    [filesByCategory]
  );

  const onChooseFile = (event: ChangeEvent<HTMLInputElement>) => {
    setSelectedFile(event.target.files?.[0] ?? null);
  };

  const onUpload = async () => {
    if (!selectedFile) {
      setError('Сначала выберите файл.');
      return;
    }

    setError(null);
    setUploading(true);
    setUploadProgress(0);

    if (backendUnsupported) {
      const objectUrl = URL.createObjectURL(selectedFile);
      setFilesByCategory((prev) => ({
        ...prev,
        [activeCategory]: [
          {
            id: `local-${Date.now()}`,
            name: selectedFile.name,
            size_bytes: selectedFile.size,
            category: activeCategory,
            uploaded_at: new Date().toISOString(),
            download_url: objectUrl,
            localObjectUrl: objectUrl,
            source: 'local-stub'
          },
          ...prev[activeCategory]
        ]
      }));
      setSelectedFile(null);
      setUploading(false);
      setUploadProgress(100);
      return;
    }

    try {
      const uploaded = await uploadFile(selectedFile, activeCategory, setUploadProgress);
      setFilesByCategory((prev) => ({
        ...prev,
        [activeCategory]: [{ ...uploaded, source: 'server' }, ...prev[activeCategory]]
      }));
      setSelectedFile(null);
    } catch (e) {
      if (isUnsupportedEndpoint(e)) {
        setBackendUnsupported(true);
        setError(
          'Файловый API на бэкенде не найден. Добавил локальную заглушку: файл будет отображаться в списке только в рамках этой сессии.'
        );
      } else {
        setError(parseError(e));
      }
    } finally {
      setUploading(false);
    }
  };

  return (
    <section className="space-y-4">
      <header className="theme-card rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm">
        <h2 className="text-lg font-semibold text-[rgb(var(--text-primary))]">Files</h2>
        <p className="mt-1 text-sm text-[rgb(var(--text-secondary))]">
          Загрузка файлов (включая большие файлы &gt; 1 GB), просмотр списком и скачивание.
        </p>

        <div className="mt-4 flex flex-wrap gap-2">
          {categories.map((category) => (
            <button
              key={category.value}
              className={
                activeCategory === category.value
                  ? 'interactive-chip rounded-lg border border-[rgb(var(--border))] bg-[rgb(var(--bg-muted))] px-3 py-1.5 text-sm font-medium text-[rgb(var(--text-primary))]'
                  : 'interactive-chip rounded-lg border border-[rgb(var(--border))] px-3 py-1.5 text-sm text-[rgb(var(--text-secondary))]'
              }
              onClick={() => setActiveCategory(category.value)}
              type="button"
            >
              {category.label}
            </button>
          ))}
        </div>

        <div className="mt-4 flex flex-wrap items-center gap-3">
          <input className="form-input max-w-md rounded-lg px-3 py-2" onChange={onChooseFile} type="file" />
          <button className="interactive-chip theme-button px-4 py-2 disabled:opacity-60" disabled={!selectedFile || uploading} onClick={() => void onUpload()} type="button">
            {uploading ? 'Загрузка…' : 'Загрузить файл'}
          </button>
        </div>

        {selectedFile && (
          <p className="mt-2 text-sm text-[rgb(var(--text-secondary))]">
            Выбран: {selectedFile.name} ({formatBytes(selectedFile.size)})
          </p>
        )}

        {(uploading || uploadProgress > 0) && (
          <div className="mt-3">
            <div className="h-2 w-full overflow-hidden rounded-full bg-[rgb(var(--bg-muted))]">
              <div className="h-full bg-blue-500 transition-all" style={{ width: `${uploadProgress}%` }} />
            </div>
            <p className="mt-1 text-xs text-[rgb(var(--text-secondary))]">{uploadProgress}%</p>
          </div>
        )}

        {error && <p className="mt-3 field-error">{error}</p>}
      </header>

      <section className="theme-card rounded-2xl border border-[rgb(var(--border))] p-5 shadow-sm">
        <div className="mb-3 flex items-center justify-between">
          <h3 className="text-base font-semibold text-[rgb(var(--text-primary))]">
            Список файлов: {activeCategory}
          </h3>
          <button className="interactive-chip rounded border border-[rgb(var(--border))] px-3 py-1 text-sm text-[rgb(var(--text-primary))]" onClick={() => void loadFiles()} type="button">
            Обновить
          </button>
        </div>

        {loading && <p className="text-sm text-[rgb(var(--text-secondary))]">Загрузка...</p>}
        {!loading && files.length === 0 && (
          <p className="text-sm text-[rgb(var(--text-secondary))]">Файлов пока нет.</p>
        )}

        <div className="space-y-3">
          {files.map((file) => (
            <article className="rounded border border-[rgb(var(--border))] p-3" key={file.id}>
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p className="font-medium text-[rgb(var(--text-primary))]">{file.name}</p>
                  <p className="text-sm text-[rgb(var(--text-secondary))]">{formatBytes(file.size_bytes)}</p>
                  <p className="text-xs text-[rgb(var(--text-secondary))]">
                    {file.uploaded_at ? new Date(file.uploaded_at).toLocaleString() : 'Дата неизвестна'}
                    {file.source === 'local-stub' ? ' · локальная заглушка' : ''}
                  </p>
                </div>
                {file.download_url ? (
                  <a className="interactive-chip rounded border border-[rgb(var(--border))] px-3 py-1 text-sm text-[rgb(var(--text-primary))]" download href={file.download_url} rel="noreferrer" target="_blank">
                    Скачать
                  </a>
                ) : (
                  <span className="text-xs text-[rgb(var(--text-secondary))]">Ссылка на скачивание недоступна</span>
                )}
              </div>
            </article>
          ))}
        </div>
      </section>
    </section>
  );
};
