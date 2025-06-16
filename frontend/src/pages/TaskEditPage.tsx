import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useApi } from '../context/api';
import { ModelsTaskByIdResponse } from '../api';
import { Loader, Upload } from 'lucide-react';
import TaskForm from '../components/TaskForm';
import { Button } from '../components/ui/button';
import useTopics from '../data/topics';

export default function TaskEditPage() {
  const { id: taskIdStr } = useParams<{ id: string }>();
  const taskId = Number(taskIdStr);
  const [task, setTask] = useState<ModelsTaskByIdResponse | undefined>(undefined);
  const [formData, setFormData] = useState({
    title: '',
    difficulty: '',
    tags: [],
    description: '',
  });
  const topics = useTopics()
  const [uploading, setUploading] = useState<{ [testCaseId: number]: number }>({});
  const [errors, setErrors] = useState<{ [testCaseId: number]: string }>({});

  const api = useApi();
  const navigate = useNavigate();

  useEffect(() => {
    if (!topics) {
        return
    }
    api.tasks.tasksIdGet({ id: taskId })
      .then((t) => {
        setTask(t);
        setFormData({
          title: t.title || '',
          difficulty: String(t.difficulty || ''),
          description: t.description || '',
          tags: (t.topics || []).map((topic) => ({
            label: topics.get(topic)?.name,
            value: topic,
          })),
        });
        return api.topics.topicsGet();
      })
      .catch(() => navigate('/'));
  }, [topics]);

  const createTestCase = async () => {
    try {
      const newCase = await api.testCases.testCasesPost({ testCase: { taskId } });
      setTask((prev) =>
        prev
          ? { ...prev, testCases: [...(prev.testCases || []), newCase] }
          : prev
      );
    } catch (err) {
      alert('Не удалось создать тест-кейс.');
    }
  };

  const handleFileUpload = (testCaseId: number, file: File, type: 'input' | 'output') => {
    const formData = new FormData();
    formData.append('file', file);
  
    const xhr = new XMLHttpRequest();
    xhr.open('POST', `http://localhost:4000/api/test-cases/${testCaseId}/upload?type=${type}`);
    const token = localStorage.getItem('token');
    if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`);
  
    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) {
        const percent = Math.round((e.loaded / e.total) * 100);
        setUploading((prev) => ({ ...prev, [`${testCaseId}_${type}`]: percent }));
      }
    };
  
    xhr.onload = () => {
      if (xhr.status === 200) {
        setErrors((prev) => ({ ...prev, [`${testCaseId}_${type}`]: '' }));
      } else {
        setErrors((prev) => ({ ...prev, [`${testCaseId}_${type}`]: `Ошибка: ${xhr.statusText}` }));
      }
      setUploading((prev) => ({ ...prev, [`${testCaseId}_${type}`]: 0 }));
    };
  
    xhr.onerror = () => {
      setErrors((prev) => ({ ...prev, [`${testCaseId}_${type}`]: 'Ошибка сети' }));
      setUploading((prev) => ({ ...prev, [`${testCaseId}_${type}`]: 0 }));
    };
  
    xhr.send(formData);
  };
  

  const handleFileChange = (
    testCaseId: number,
    type: 'input' | 'output',
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (file.size > 20 * 1024 * 1024) {
      alert('Файл больше 20 МБ');
      return;
    }
    handleFileUpload(testCaseId, file, type);
  };
  

  if (!task) {
    return (
      <div className="flex items-center justify-center">
        <Loader className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto p-6 space-y-8">
      <h2 className="text-3xl font-bold">Редактирование задачи #{taskId}</h2>

      <TaskForm formData={formData} setFormData={setFormData} topicsList={Array.from(topics?.values()!, (k ,v) => ({label: k.name!, value: v}))} />

      <div className="space-y-4">
        <div className="flex justify-between items-center">
          <h3 className="text-xl font-semibold">Тестовые сценарии</h3>
          <Button onClick={createTestCase}>
            <Upload className="w-4 h-4 mr-2" /> Добавить сценарий
          </Button>
        </div>

        {task.testCases?.length ? (
          <ul className="space-y-4">
            {task.testCases.map((tc) => (
              <li key={tc.id} className="border rounded-md p-4 space-y-2">
                <div className="flex justify-between items-center">
                  <span className="font-medium">Тест #{tc.caseOrder+1}</span>
                  <div className="space-y-2">
                    <div>
                      <label className="block text-sm font-medium">Входной файл</label>
                      <input
                        type="file"
                        onChange={(e) => handleFileChange(tc.id!, 'input', e)}
                        className="text-sm"
                      />
                      {uploading[`${tc.id}_input`] > 0 && (
                        <div className="w-full bg-gray-700 rounded-full overflow-hidden h-2 mt-1">
                          <div
                            className="bg-brand h-2"
                            style={{ width: `${uploading[`${tc.id}_input`]}%` }}
                          />
                        </div>
                      )}
                      {errors[`${tc.id}_input`] && (
                        <div className="text-sm text-red-500">{errors[`${tc.id}_input`]}</div>
                      )}
                    </div>

                    <div>
                      <label className="block text-sm font-medium">Выходной файл</label>
                      <input
                        type="file"
                        onChange={(e) => handleFileChange(tc.id!, 'output', e)}
                        className="text-sm"
                      />
                      {uploading[`${tc.id}_output`] > 0 && (
                        <div className="w-full bg-gray-700 rounded-full overflow-hidden h-2 mt-1">
                          <div
                            className="bg-brand h-2"
                            style={{ width: `${uploading[`${tc.id}_output`]}%` }}
                          />
                        </div>
                      )}
                      {errors[`${tc.id}_output`] && (
                        <div className="text-sm text-red-500">{errors[`${tc.id}_output`]}</div>
                      )}
                    </div>
                  </div>
                </div>
                {uploading[tc.id!] > 0 && (
                  <div className="w-full bg-gray-700 rounded-full overflow-hidden h-2">
                    <div
                      className="bg-brand h-2"
                      style={{ width: `${uploading[tc.id!]}%` }}
                    />
                  </div>
                )}
                {errors[tc.id!] && (
                  <div className="text-sm text-red-500">{errors[tc.id!]}</div>
                )}
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-muted-foreground">Для этой задачи еще не создано тестовых сценариев</p>
        )}
      </div>
    </div>
  );
}
