import { DbGetUsersLeaderboardRow } from "../api";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../components/ui/table";
import { useApi } from "../context/api";
import React, { useEffect, useState } from 'react';

export default function Leaderboard() {
    const api = useApi()
    const [users, setUsers] = useState<DbGetUsersLeaderboardRow[]>(new Array());
    useEffect(() => {
        api.users.leaderboardGet().then((users) => {
            setUsers(users)
        })
    }, [])
     return (
        <>
        <h2 className="text-center p-4">Таблица лидеров</h2>
        <div className="rounded border flex w-[50%] m-auto">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Имя</TableHead>
                        <TableHead>Elo</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {users.map((u) => (
                        <TableRow key={u.username}>
                            <TableCell>{u.username}</TableCell>
                            <TableCell>{u.elo}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
        </>
     )
}