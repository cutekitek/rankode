import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/button";
import { useApi } from "../context/api";
import TaskForm from "../components/TaskForm";
import React from 'react';

export default function CreateTaskPage() {
  const api = useApi();
  const navigate = useNavigate();

  const [formData, setFormData] = useState({
    title: '',
    difficulty: -1,
    tags: [] as any[],
    description: '',
  });

  const [topicsList, setTopicsList] = useState<any[]>([]);

  useEffect(() => {
    api.topics.topicsGet().then((res) => {
      const options = res.map((topic) => ({
        label: topic.name,
        value: topic.id,
      }));
      setTopicsList(options);
    });
  }, []);

  const handleCreate = async () => {
    try {
      const task = await api.tasks.tasksPost({
        task: {
          title: formData.title,
          description: formData.description,
          difficulty: formData.difficulty,
          topics: formData.tags.map((t) => t.value),
        }
      });
      navigate(`/task/${task.id}/edit`);
    } catch (err) {
      console.error(err);
      alert('Failed to create task');
    }
  };

  return (
    <div className="max-w-3xl mx-auto p-6 space-y-6">
      <h2 className="text-3xl font-bold">Создать задачу</h2>

      <TaskForm
        formData={formData}
        setFormData={setFormData}
        topicsList={topicsList}
      />

      <div>
        <Button onClick={handleCreate} className="w-full">Создать</Button>
      </div>
    </div>
  );
}
