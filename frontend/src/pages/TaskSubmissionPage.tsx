import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { ResizablePanelGroup, ResizablePanel, ResizableHandle } from '../components/ui/resizable';
import  Editor from 'react-simple-code-editor';
import { highlight, languages } from 'prismjs';
import 'prismjs/components/prism-clike';
import 'prismjs/components/prism-javascript';
import 'prismjs/components/prism-python';
import 'prismjs/components/prism-go'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../components/ui/table';
import { Loader } from 'lucide-react';
import { Button } from '../components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { toast } from "sonner"
import { useApi } from '../context/api';


import "../themes/gruvbox.css";
import { ModelsGetUserTaskAttemptsResponse } from '../api';

interface Task {
  id: number;
  title: string;
  description: string;
  difficulty: number;
  topics: number[];
}

interface Topic {
  id: number;
  name: string;
}

const languageOptions = [
  { value: 'javascript', label: 'JavaScript' },
  { value: 'python3', label: 'Python' },
  { value: 'cpp', label: 'C++' },
  { value: 'go', label: 'Golang'}
];

const languageGrammarMap = {
  "python3": "python",
}

const statusMap: Record<number, { label: string; color: string }> = {
  0: { label: 'Успешно', color: 'bg-green-100 text-green-800' },
  1: { label: 'Ошибка компиляции', color: 'bg-red-100 text-red-800' },
  2: { label: 'Ошибка запуска', color: 'bg-red-100 text-red-800' },
  3: { label: 'Внутренняя ошибка', color: 'bg-red-100 text-red-800' },
  4: { label: 'Обрабатывается', color: 'bg-yellow-100 text-yellow-800' },
  5: { label: 'Неправильный ответ', color: 'bg-red-100 text-red-800' },
};

export default function TaskSubmissionPage() {
  const { id: taskId } = useParams<{ id: string }>();
  const api = useApi();
  
  const [task, setTask] = useState<Task | null>(null);
  const [topics, setTopics] = useState<Record<number, Topic>>({});
  const [attempts, setAttempts] = useState<ModelsGetUserTaskAttemptsResponse[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  const [code, setCode] = useState('');
  const [selectedLanguage, setSelectedLanguage] = useState('javascript');
  const [codeLoaded, setCodeLoaded] = useState(false)

  // Загрузка данных задачи и попыток
  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);
        
        // Загрузка задачи
        const taskData = await api.tasks.tasksIdGet({ id: parseInt(taskId!) });
        setTask(taskData);
        
        // Загрузка тем
        const topicsResponse = await api.topics.topicsGet();
        const topicsMap = topicsResponse.reduce((acc: Record<number, Topic>, topic) => {
          acc[topic.id] = topic;
          return acc;
        }, {});
        setTopics(topicsMap);
        
        // Загрузка попыток
        const attemptsData = await api.attempts.attemptsGet({ taskId: parseInt(taskId!) });
        setAttempts(attemptsData);        
      } catch (error) {
        toast.error('Ошибка загрузки', {
          description: 'Не удалось загрузить данные задачи',
        });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [taskId, api]);

  useEffect(() => {
    if (!isLoading && !codeLoaded) {
      for (let a of attempts) {
        console.log(a);
          if (a.status == 0 || a.status == 4) {
            console.log(a)
            setCode(a.code!)
            setSelectedLanguage(a.lang || languageOptions[0].value);
            break
          }
      }
      setCodeLoaded(true);
    }
  }, [attempts, codeLoaded])

  // Обработка отправки решения
  const handleSubmit = async () => {
    if (!code.trim()) {
      return;
    }

    try {
      setIsSubmitting(true);
      
      await api.attempts.attemptsPost({
        attempt: {
          code: code,
          lang: selectedLanguage,
          taskId: parseInt(taskId!),
        }
      });

      toast('Решение отправлено', {
        description: 'Ваше решение успешно отправлено на проверку',
      });
      
      // Обновляем список попыток
      const updatedAttempts = await api.attempts.attemptsGet({ taskId: parseInt(taskId!) });
      setAttempts(updatedAttempts);
      
    } catch (error) {
        toast.error('Ошибка отпраки', {
            description: 'Не удалось отправить задачу',
        });
    } finally {
      setIsSubmitting(false);
    }
  };

  // Форматирование даты
  const formatDate = (dateString: string) => {
    return (new Date(dateString)).toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (isLoading || !task) {
    return (
      <div className="flex items-center justify-center h-screen">
        <Loader className="h-12 w-12 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="h-screen flex flex-col">
      <div className="p-4 border-b">
        <h1 className="text-xl font-bold">{task.title}</h1>
        <div className="flex flex-wrap gap-2 mt-2">
          <span className={`px-2 py-1 rounded text-xs ${
            task.difficulty === 0 ? 'bg-green-100 text-green-800' :
            task.difficulty === 1 ? 'bg-yellow-100 text-yellow-800' : 'bg-red-100 text-red-800'
          }`}>
            {task.difficulty === 0 ? 'Легко' : task.difficulty === 1 ? 'Средне' : 'Сложно'}
          </span>
          {task.topics.map(topicId => (
            <span 
              key={topicId} 
              className="px-2 py-1 rounded bg-blue-100 text-blue-800 text-xs"
            >
              {topics[topicId]?.name || `Тема ${topicId}`}
            </span>
          ))}
        </div>
      </div>

      <ResizablePanelGroup direction="horizontal" className="flex-grow">
        {/* Левый столбец */}
        <ResizablePanel defaultSize={50} minSize={30}>
          <ResizablePanelGroup direction="vertical">
            <ResizablePanel defaultSize={70} minSize={30}>
              <div className="h-full p-4 overflow-auto">
                <h2 className="font-semibold mb-3">Описание задачи</h2>
                <div className="prose prose-sm max-w-none">
                  {task.description.split('\n').map((paragraph, index) => (
                    <p key={index} className="mb-3">{paragraph}</p>
                  ))}
                </div>
              </div>
            </ResizablePanel>
            
            <ResizableHandle withHandle />
            
            <ResizablePanel defaultSize={30} minSize={20}>
              <div className="h-full flex flex-col">
                <div className="p-4 border-b">
                  <h2 className="font-semibold">История попыток</h2>
                </div>
                <div className="flex-grow overflow-auto">
                  {attempts.length === 0 ? (
                    <div className="flex items-center justify-center h-full text-muted-foreground">
                      Пока нет попыток
                    </div>
                  ) : (
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>Дата</TableHead>
                          <TableHead>Язык</TableHead>
                          <TableHead>Статус</TableHead>
                          <TableHead>Время</TableHead>
                          <TableHead>Память</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {attempts.map((attempt) => (
                          <TableRow key={attempt.id}>
                            <TableCell className="py-2 px-4">
                              {formatDate(attempt.updatedAt!)}
                            </TableCell>
                            <TableCell className="py-2 px-4">
                              {attempt.lang}
                            </TableCell>
                            <TableCell className="py-2 px-4">
                              <span className={`px-2 py-1 rounded text-xs ${
                                statusMap[attempt.status!]?.color || 'bg-gray-100 text-gray-800'
                              }`}>
                                {statusMap[attempt.status!]?.label || 'Неизвестно'}
                              </span>
                            </TableCell>
                            <TableCell className="py-2 px-4">
                              {attempt.runningTime ? `${attempt.runningTime} ms` : '-'}
                            </TableCell>
                            <TableCell className="py-2 px-4">
                              {attempt.memoryUsage ? `${attempt.memoryUsage} kb` : '-'}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  )}
                </div>
              </div>
            </ResizablePanel>
          </ResizablePanelGroup>
        </ResizablePanel>

        <ResizableHandle withHandle />

        {/* Правый столбец - редактор кода */}
        <ResizablePanel defaultSize={50} minSize={40}>
          <div className="h-full flex flex-col">
            <div className="p-2 border-b flex justify-between items-center">
              <h2 className="font-semibold">Редактор кода</h2>
              <div className="flex items-center gap-4">
                <Select 
                  value={selectedLanguage}
                  onValueChange={setSelectedLanguage}
                >
                  <SelectTrigger className="w-[120px]">
                    <SelectValue placeholder="Язык" />
                  </SelectTrigger>
                  <SelectContent>
                    {languageOptions.map((lang) => (
                      <SelectItem key={lang.value} value={lang.value}>
                        {lang.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Button 
                  onClick={handleSubmit}
                  disabled={isSubmitting}
                >
                  {isSubmitting ? (
                    <>
                      <Loader className="mr-2 h-4 w-4 animate-spin" />
                      Отправка...
                    </>
                  ) : 'Отправить'}
                </Button>
              </div>
            </div>
            
            <div className="flex-grow overflow-hidden">
              <Editor
                value={code}
                onValueChange={setCode}
                highlight={(code) => highlight(code, languages[languageGrammarMap[selectedLanguage] || selectedLanguage], selectedLanguage)}
                padding={15}
                style={{
                  height: '100%',
                  overflow: 'auto',
                  backgroundColor: '#111b27',
                }}
                textareaClassName="focus:outline-none"
              />
            </div>
          </div>
        </ResizablePanel>
      </ResizablePanelGroup>
    </div>
  );
}