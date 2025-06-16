import { useEffect, useMemo, useState } from "react";
import { DbTopic } from "../api/models";
import { useApi } from "../context/api";

export default function useTopics() {
    const api = useApi();
    const [topicsList, setTopicsList] = useState<Map<number, DbTopic> | null>(null)
    useEffect(() => {
        api.topics.topicsGet().then((t) => {
            const topics = new Map<number, DbTopic>();
            t.forEach(e => topics.set(e.id!, e))
            setTopicsList(topics)
        }
        )
    }, [])
    return topicsList
}
